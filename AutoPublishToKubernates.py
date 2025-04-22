# -*- coding: utf-8 -*-
import os
import subprocess
import sys
import shutil
import platform
import socket
import time
import webbrowser
import signal # Added for graceful exit handling

# --- Configuration Templates ---

DOCKERFILE_TEMPLATE = '''# Use the official Golang image to create a build artifact.
# https://hub.docker.com/_/golang
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy local code to the container image.
COPY . .

# Build the binary.
# -ldflags="-s -w" strips debug information and symbols, reducing binary size.
# CGO_ENABLED=0 prevents CGo usage, ensuring static linking (usually).
# GOOS=linux GOARCH=amd64 explicitly sets the target OS/Arch for the container.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main .

# Use a slim Debian image for the final stage.
# https://hub.docker.com/_/debian
FROM debian:bullseye-slim

# Install necessary packages:
# ca-certificates: For HTTPS connections
# curl: Useful for debugging within the container
# iputils-ping: For network diagnostics (ping)
# dnsutils: For DNS lookups (dig, nslookup)
# net-tools: Network utilities (netstat, etc.)
# bash: Common shell
# Clean up apt cache to reduce image size.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    iputils-ping \
    dnsutils \
    net-tools \
    bash \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from the builder stage.
COPY --from=builder /app/main /app/main

# Grant execute permission to the binary
RUN chmod +x /app/main

# Expose the port the application listens on.
EXPOSE {port}

# Command to run the executable.
CMD ["/app/main"]
'''

DEPLOYMENT_YAML_TEMPLATE = '''apiVersion: apps/v1
kind: Deployment
metadata:
  name: {name}-deployment
  labels:
    app: {name}
spec:
  replicas: 1 # Start with one instance
  selector:
    matchLabels:
      app: {name} # Selects pods with this label
  template:
    metadata:
      labels:
        app: {name} # Labels pods created by this deployment
    spec:
      containers:
      - name: {name}-container
        image: {image_name} # Use the image built by the script
        imagePullPolicy: IfNotPresent # Only pull if not locally present (important for kind)
        ports:
        - containerPort: {port} # Port the container listens on
        envFrom: # Load environment variables from a Secret
        - secretRef:
            name: {secret_name} # Name of the secret to load from
'''

SERVICE_YAML_TEMPLATE = '''apiVersion: v1
kind: Service
metadata:
  name: {name}-service
  labels:
    app: {name}
spec:
  type: NodePort # Exposes the Service on each Node's IP at a static port (the NodePort).
                 # For local 'kind', port-forwarding is often easier for access.
  selector:
    app: {name} # Selects pods with this label to route traffic to
  ports:
    - name: http # Name for the port (optional but good practice)
      protocol: TCP
      port: {port} # Port the service listens on internally within the cluster
      targetPort: {port} # Port on the Pods to forward traffic to
      # nodePort: 30001 # Optional: Specify a static NodePort (range 30000-32767)
                      # If omitted, Kubernetes assigns a random one.
                      # Let's omit it for more flexibility, port-forward is preferred.
'''

# Global variable to hold the port-forward process
port_forward_process = None

# --- Helper Functions ---

def run_command(command, cwd=None, check=True, capture_output=False, text=True):
    """Helper function to run shell commands."""
    print(f"🔧 Running command: {' '.join(command)}{f' in {cwd}' if cwd else ''}")
    try:
        result = subprocess.run(
            command,
            cwd=cwd,
            check=check,
            capture_output=capture_output,
            text=text,
            # Use shell=True only if the command string needs shell interpretation (e.g., pipes, wildcards)
            # For list-based commands, shell=False is safer.
            shell=isinstance(command, str)
        )
        if capture_output and result.stdout:
            print(f" L_ Output: {result.stdout.strip()}")
        if result.stderr:
             print(f" L_ Error Output: {result.stderr.strip()}", file=sys.stderr)
        return result
    except subprocess.CalledProcessError as e:
        print(f"❌ Command failed: {' '.join(command) if isinstance(command, list) else command}", file=sys.stderr)
        # Captured output is available on the exception object
        if e.stdout:
            print(f"   stdout: {e.stdout}", file=sys.stderr)
        if e.stderr:
            print(f"   stderr: {e.stderr}", file=sys.stderr)
        sys.exit(1)
    except FileNotFoundError:
        print(f"❌ Command not found: {command[0]}. Is it installed and in PATH?", file=sys.stderr)
        sys.exit(1)


def check_and_install_tool(tool_name, install_commands):
    """Checks if a tool exists, installs it if not."""
    if shutil.which(tool_name):
        print(f"✅ {tool_name} is already installed.")
        return

    print(f"🤔 {tool_name} not found. Attempting installation...")
    for command_str in install_commands:
        print(f"   -> Running: {command_str}")
        # Use shell=True here because install commands often use pipes, sudo etc.
        try:
            # Use Popen to stream output in real-time
            process = subprocess.Popen(command_str, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, bufsize=1, universal_newlines=True)
            for line in process.stdout:
                 print(f"      {line.strip()}", flush=True) # Stream install output
            process.wait() # Wait for the command to complete
            rc = process.poll()
            if rc != 0:
                raise subprocess.CalledProcessError(rc, command_str)
        except subprocess.CalledProcessError as e:
            print(f"❌ Failed to install {tool_name} with command: {command_str}", file=sys.stderr)
            print(f"   Return code: {e.returncode}", file=sys.stderr)
            # Output was already streamed
            sys.exit(1)
        except FileNotFoundError:
             print(f"❌ Error running install command '{command_str}'. Is a required tool (like sudo, curl, apt-get, brew) missing or not in PATH?", file=sys.stderr)
             sys.exit(1)

    # Verify installation
    if shutil.which(tool_name):
        print(f"✅ Successfully installed {tool_name}.")
    else:
        print(f"❌ Installation command ran, but {tool_name} still not found in PATH. Please check installation steps and PATH configuration.", file=sys.stderr)
        sys.exit(1)


def is_port_in_use(port, host='127.0.0.1'):
    """Checks if a specific host/port combination is currently in use."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        try:
            # Try to bind to the specific host and port
            s.bind((host, port))
            return False # Bind succeeded, port is likely free on this interface
        except socket.error:
            return True # Bind failed, port is likely in use on this interface

def get_valid_port():
    """Prompts user for a port and validates it."""
    default_port = 8080
    while True:
        port_input = input(f"🔌 Enter the application port (leave empty for default {default_port}): ").strip()
        if not port_input:
            port = default_port
        elif not port_input.isdigit():
            print("🚫 Invalid input. Port must be a number.", file=sys.stderr)
            continue
        else:
            port = int(port_input)
            if not 1 <= port <= 65535:
                 print(f"🚫 Invalid port number ({port}). Must be between 1 and 65535.", file=sys.stderr)
                 continue

        # Check if port is potentially used by the script itself or common services
        # Check both 127.0.0.1 and 0.0.0.0 because we'll bind to 0.0.0.0 later
        if is_port_in_use(port, '127.0.0.1') or is_port_in_use(port, '0.0.0.0'):
            print(f"⚠️ Port {port} seems to be in use locally (on 127.0.0.1 or 0.0.0.0). This WILL conflict with port-forwarding. Please choose a different port.")
            continue # Force user to choose another port if already bound
        else:
            return port # Port seems free

def write_files(folder_name, port, image_name, secret_name):
    """Generates Dockerfile, Kubernetes manifests, and default main.go if needed."""
    print(f"📝 Writing configuration files in '{folder_name}'...")
    go_file_path = os.path.join(folder_name, "main.go")

    # Create default main.go if it doesn't exist
    if not os.path.isfile(go_file_path):
        print(f"   -> File '{go_file_path}' not found. Creating a default Go server.")
        app_name = os.path.basename(folder_name)
        default_go_code = f'''package main

import (
	"fmt"
	"log"
	"net/http"
	"os" // Import os package
	"time" // Import time package
)

func main() {{
	// Use the PORT environment variable if set, otherwise default
	port := os.Getenv("PORT")
	if port == "" {{
		port = "{port}" // Default port passed from the script
	}}

	// Get environment variable example (replace with your actual env var name)
	secretValue := os.Getenv("MY_SECRET_KEY") // Example: Get env var from secret
	if secretValue == "" {{
		secretValue = "(not set)"
	}}

	hostname, _ := os.Hostname() // Get hostname for logging

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {{
		startTime := time.Now()
		// Example of using an environment variable from the secret
		fmt.Fprintf(w, "✅ Hello from {app_name} on host '%s'!\\n", hostname)
		fmt.Fprintf(w, "   Port: %s\\n", port)
		fmt.Fprintf(w, "   Secret Value (MY_SECRET_KEY): %s\\n", secretValue)
		fmt.Fprintf(w, "   Client Address: %s\\n", r.RemoteAddr)
		fmt.Fprintf(w, "   Request Path: %s\\n", r.URL.Path)
		log.Printf("Request from %s for %s (served by %s) - Took %v", r.RemoteAddr, r.URL.Path, hostname, time.Since(startTime))
	}})

	listenAddr := fmt.Sprintf(":%s", port)
	fmt.Printf("🚀 Starting server for '{app_name}' on host '%s'\\n", hostname)
	fmt.Printf("👂 Listening on internal address %s\\n", listenAddr)
	fmt.Printf("🤫 Example Secret 'MY_SECRET_KEY' is: %s\\n", secretValue) // Show on startup too

	// Use log.Fatal for server errors
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}}
'''
        with open(go_file_path, "w", encoding="utf-8") as f:
            f.write(default_go_code)
        print(f"   -> Default 'main.go' created.")

    # Write Dockerfile
    dockerfile_path = os.path.join(folder_name, "Dockerfile")
    with open(dockerfile_path, "w", encoding="utf-8") as f:
        f.write(DOCKERFILE_TEMPLATE.format(port=port))
    print(f"   -> 'Dockerfile' written.")

    # Write Kubernetes Deployment YAML
    deployment_yaml_path = os.path.join(folder_name, "deployment.yaml")
    app_name = os.path.basename(folder_name).lower().replace("_", "-") # Use folder name as app name, ensure k8s compliant
    with open(deployment_yaml_path, "w", encoding="utf-8") as f:
        f.write(DEPLOYMENT_YAML_TEMPLATE.format(name=app_name, port=port, image_name=image_name, secret_name=secret_name))
    print(f"   -> 'deployment.yaml' written.")

    # Write Kubernetes Service YAML
    service_yaml_path = os.path.join(folder_name, "service.yaml")
    with open(service_yaml_path, "w", encoding="utf-8") as f:
        f.write(SERVICE_YAML_TEMPLATE.format(name=app_name, port=port))
    print(f"   -> 'service.yaml' written.")

    return app_name # Return the derived application name

def image_exists_locally(image_name):
    """Checks if a Docker image exists locally."""
    print(f"[*] Checking for local Docker image: {image_name}...")
    result = run_command(["docker", "images", "-q", image_name], check=False, capture_output=True)
    exists = result.stdout.strip() != ""
    print(f"   -> Exists: {exists}")
    return exists

def build_image(folder_name, image_name):
    """Builds the Docker image."""
    print(f"[*] Building Docker image: {image_name} from '{folder_name}'...")
    # Use run_command helper
    run_command(["docker", "build", "-t", image_name, "."], cwd=folder_name)
    print(f"✅ Docker image '{image_name}' built successfully.")

def ensure_kind_cluster(cluster_name="go-cluster"):
    """Ensures the specified kind cluster is running."""
    print(f"[*] Checking kind cluster '{cluster_name}' status...")
    try:
        result = run_command(["kind", "get", "clusters"], capture_output=True, check=False) # Don't exit if kind fails temporarily
        if result.returncode != 0:
             print(f"   -> Failed to get kind clusters (is kind installed and configured?). Attempting to create...")
             run_command(["kind", "create", "cluster", f"--name={cluster_name}"])
             print(f"✅ kind cluster '{cluster_name}' created.")
             return True

        clusters = result.stdout.splitlines()
        if cluster_name in clusters:
            print(f"✅ kind cluster '{cluster_name}' is already running.")
            return True
        else:
            print(f"   -> Cluster '{cluster_name}' not found. Creating...")
            run_command(["kind", "create", "cluster", f"--name={cluster_name}"])
            print(f"✅ kind cluster '{cluster_name}' created.")
            return True
    except FileNotFoundError:
         print("❌ 'kind' command not found. Please ensure kind is installed and in your PATH.", file=sys.stderr)
         sys.exit(1)
    except subprocess.CalledProcessError:
        # Error messages handled by run_command
        print(f"❌ Failed to ensure kind cluster '{cluster_name}' is running.", file=sys.stderr)
        return False # Indicate failure

def load_image_to_kind(image_name, cluster_name="go-cluster"):
    """Loads the Docker image into the kind cluster."""
    print(f"[*] Loading image '{image_name}' into kind cluster '{cluster_name}'...")
    try:
        run_command(["kind", "load", "docker-image", image_name, f"--name={cluster_name}"])
        print(f"✅ Image '{image_name}' loaded into '{cluster_name}'.")
    except FileNotFoundError:
         print("❌ 'kind' command not found. Please ensure kind is installed and in your PATH.", file=sys.stderr)
         sys.exit(1)
    except subprocess.CalledProcessError:
        # Error messages handled by run_command
        print(f"❌ Failed to load image '{image_name}' into kind cluster '{cluster_name}'.", file=sys.stderr)
        print("   -> Hint: Make sure the Docker daemon is running and the image exists locally ('docker images').", file=sys.stderr)
        sys.exit(1)


def apply_k8s_resources(folder_name):
    """Applies Kubernetes deployment and service manifests."""
    print("[*] Applying Kubernetes resources...")
    try:
        print("   -> Applying Deployment...")
        run_command(["kubectl", "apply", "-f", "deployment.yaml"], cwd=folder_name)
        print("   -> Applying Service...")
        run_command(["kubectl", "apply", "-f", "service.yaml"], cwd=folder_name)
        print("✅ Kubernetes resources applied.")
    except FileNotFoundError:
         print("❌ 'kubectl' command not found. Please ensure kubectl is installed and in your PATH.", file=sys.stderr)
         sys.exit(1)
    except subprocess.CalledProcessError:
        # Error messages handled by run_command
        print("❌ Failed to apply Kubernetes resources.", file=sys.stderr)
        print("   -> Hint: Ensure kubectl is configured correctly and can connect to your kind cluster ('kubectl get nodes').", file=sys.stderr)
        sys.exit(1)

def wait_for_deployment(app_name, namespace="default", timeout_secs=180):
    """Waits for a Kubernetes deployment to become ready."""
    deployment_name = f"{app_name}-deployment"
    print(f"[*] Waiting for deployment '{deployment_name}' to be ready (max {timeout_secs}s)...")
    command = [
        "kubectl", "wait",
        "--for=condition=available",
        f"deployment/{deployment_name}",
        f"--namespace={namespace}",
        f"--timeout={timeout_secs}s"
    ]
    try:
        run_command(command)
        print(f"✅ Deployment '{deployment_name}' is ready!")
        return True
    except FileNotFoundError:
         print("❌ 'kubectl' command not found. Cannot check deployment status.", file=sys.stderr)
         sys.exit(1)
    except subprocess.CalledProcessError:
        # Error messages handled by run_command
        print(f"❌ Deployment '{deployment_name}' did not become ready within the timeout.", file=sys.stderr)
        print(f"   -> Check pod status: 'kubectl get pods -l app={app_name} -n {namespace}'")
        print(f"   -> Check pod logs: 'kubectl logs -l app={app_name} -n {namespace}'")
        print(f"   -> Describe deployment: 'kubectl describe deployment {deployment_name} -n {namespace}'")
        print(f"   -> Describe pods: 'kubectl describe pods -l app={app_name} -n {namespace}'")
        return False # Indicate failure

def ensure_docker_running():
    """Checks if the Docker daemon is running."""
    print("[*] Checking if Docker daemon is running...")
    try:
        run_command(["docker", "info"], capture_output=True) # Use capture_output to hide verbose output
        print("✅ Docker daemon is running.")
    except FileNotFoundError:
        print("❌ 'docker' command not found. Please install Docker.", file=sys.stderr)
        sys.exit(1)
    except subprocess.CalledProcessError:
        # Error messages handled by run_command
        print("❌ Docker daemon does not seem to be running.", file=sys.stderr)
        print("   -> Please start the Docker daemon and try again.", file=sys.stderr)
        sys.exit(1)

def install_requirements():
    """Installs required tools based on the operating system."""
    print("[*] Checking and installing required tools...")
    system = platform.system()

    if system == "Linux":
        # Check for sudo
        if os.geteuid() != 0:
             print("   -> Linux detected. Some install commands may require root privileges (sudo).")
             # Check if sudo is available
             if not shutil.which("sudo"):
                  print("❌ 'sudo' command not found, but root privileges are likely needed. Please run as root or install sudo.", file=sys.stderr)
                  # sys.exit(1) # Or allow to continue and potentially fail later
        # Docker
        check_and_install_tool("docker", [
            "sudo apt-get update",
            "sudo apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release", # Docker prereqs
            "sudo install -m 0755 -d /etc/apt/keyrings", # Ensure keyring dir exists
            "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg", # Use Docker official repo
            "sudo chmod a+r /etc/apt/keyrings/docker.gpg",
            'echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null',
            "sudo apt-get update",
            "sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin", # Install Docker engine & tools
            "sudo systemctl start docker", # Attempt to start docker
            "sudo systemctl enable docker", # Enable docker on boot
            "sudo usermod -aG docker $USER || echo 'Could not add user to docker group. Run docker with sudo or configure group manually. You may need to log out and back in for group changes to take effect.'" # Add user to docker group
            ])
        # Kubectl
        check_and_install_tool("kubectl", [
            "sudo apt-get update",
            "sudo apt-get install -y apt-transport-https ca-certificates curl",
            'curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg', # Use modern apt key method (Update version string v1.30 as needed)
            'echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list',
            "sudo apt-get update",
            "sudo apt-get install -y kubectl"
            # "sudo apt-mark hold kubectl" # Optional: Prevent accidental upgrades via apt
            ])
        # Kind
        check_and_install_tool("kind", [
            'curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-amd64', # Use a specific version for stability
            'chmod +x ./kind',
            'sudo mv ./kind /usr/local/bin/kind'
            ])

    elif system == "Darwin": # macOS
        # Check for Homebrew first
        if not shutil.which("brew"):
            print("❌ Homebrew ('brew') not found. Please install it from https://brew.sh/", file=sys.stderr)
            sys.exit(1)
        # Docker Desktop (Cask)
        check_and_install_tool("docker", ["brew install --cask docker"]) # Assumes Docker Desktop
        # Kubectl
        check_and_install_tool("kubectl", ["brew install kubectl"])
        # Kind
        check_and_install_tool("kind", ["brew install kind"])
        print("   -> On macOS, ensure Docker Desktop is running.")

    elif system == "Windows":
         print("🚧 Windows support is experimental and requires manual setup.")
         print("   Please ensure Docker Desktop, kubectl, and kind are installed and in your PATH.")
         print("   Installation guides:")
         print("     - Docker Desktop: https://www.docker.com/products/docker-desktop/")
         print("     - kubectl (via Docker Desktop): Enable Kubernetes in Docker Desktop settings.")
         print("     - kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installation (use Chocolatey or download binary)")
         # Attempt basic checks
         if not shutil.which("docker"): print("   -> 'docker' not found in PATH.")
         if not shutil.which("kubectl"): print("   -> 'kubectl' not found in PATH.")
         if not shutil.which("kind"): print("   -> 'kind' not found in PATH.")
         input("   Press Enter to continue if you have installed the tools manually, or Ctrl+C to exit...")

    else:
        print(f"🚫 Unsupported operating system: {system}", file=sys.stderr)
        sys.exit(1)

def create_k8s_secret_from_env(folder_name, secret_name, namespace="default"):
    """Creates or updates a Kubernetes secret from a .env file."""
    env_path = os.path.join(folder_name, ".env")
    if not os.path.isfile(env_path):
        print(f"ℹ️ No '.env' file found in '{folder_name}'. Skipping Secret creation/update.")
        # Check if secret already exists, maybe from a previous run
        check_secret_cmd = ["kubectl", "get", "secret", secret_name, f"--namespace={namespace}", "-o", "name"]
        try:
             # Run command but don't check=True, capture output to see if it exists
             result = run_command(check_secret_cmd, capture_output=True, check=False)
             if result.returncode == 0 and result.stdout.strip():
                 print(f"   -> Note: Secret '{secret_name}' exists but no .env file provided to update it.")
        except FileNotFoundError:
             print("❌ 'kubectl' command not found. Cannot check for existing secret.", file=sys.stderr)
             # Don't exit here, maybe deployment doesn't strictly need the secret yet.
        # Don't need to handle CalledProcessError explicitly here, failure means it doesn't exist or kubectl failed.
        return

    print(f"[*] Managing Kubernetes Secret '{secret_name}' from '{env_path}'...")
    # Delete the secret if it exists (simplest way to update)
    delete_cmd = ["kubectl", "delete", "secret", secret_name, f"--namespace={namespace}", "--ignore-not-found=true"]
    try:
        run_command(delete_cmd, check=False) # Don't fail if it wasn't found

        # Create the secret from the .env file
        create_cmd = ["kubectl", "create", "secret", "generic", secret_name,
                      f"--from-env-file={env_path}", f"--namespace={namespace}"]
        run_command(create_cmd)
        print(f"✅ Secret '{secret_name}' created/updated from '{env_path}'.")

    except FileNotFoundError:
         print("❌ 'kubectl' command not found. Cannot manage secret.", file=sys.stderr)
         sys.exit(1)
    except subprocess.CalledProcessError:
        # Error message handled by run_command
        print(f"❌ Failed to manage Kubernetes secret '{secret_name}'.", file=sys.stderr)
        print(f"   -> Ensure '{env_path}' is formatted correctly (KEY=VALUE pairs, no spaces around '=').", file=sys.stderr)
        sys.exit(1)

def start_port_forward(service_name, port, namespace="default"):
    """Starts kubectl port-forward in the background, listening on all interfaces."""
    global port_forward_process
    if port_forward_process and port_forward_process.poll() is None:
        print("ℹ️ Port-forward process already seems to be running.")
        return True # Assume it's okay

    print(f"[*] Starting port-forward for svc/{service_name} on port {port} (accessible from network)...")
    command = [
        "kubectl", "port-forward",
        f"svc/{service_name}",
        f"{port}:{port}",           # Map local port to service port
        f"--namespace={namespace}",
        "--address=0.0.0.0"         # <-- Listen on all interfaces
    ]
    try:
        # Use Popen to run in the background
        print(f"   -> Executing: {' '.join(command)}")
        # Redirect stdout/stderr to pipes so we can check for immediate errors
        port_forward_process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        print(f"✅ Port-forward process started (PID: {port_forward_process.pid}).")

        # Give it a moment to establish connection or fail
        time.sleep(4) # Increased sleep slightly

        # Check if the process started successfully but exited immediately
        if port_forward_process.poll() is not None:
             print(f"❌ Port-forward process exited unexpectedly (Code: {port_forward_process.returncode}).", file=sys.stderr)
             stdout, stderr = port_forward_process.communicate() # Get output after exit
             if stdout: print(f"   stdout: {stdout.strip()}", file=sys.stderr)
             if stderr: print(f"   stderr: {stderr.strip()}", file=sys.stderr)
             print(f"   -> Check if the service 'svc/{service_name}' and its pods are running correctly ('kubectl get pods -l app={service_name.replace('-service','')})'.", file=sys.stderr)
             print(f"   -> Also check if local port {port} is already in use by another application ('netstat', 'lsof').", file=sys.stderr)
             port_forward_process = None # Reset global var
             return False # Indicate failure

        # If process is still running after sleep, assume it's working
        print(f"   -> Forwarding http://<your-pc-ip>:{port} (and localhost) -> svc/{service_name}:{port}") # MODIFIED LINE
        print("   ⚠️ WARNING: Service is now exposed to your local network.") # ADDED LINE
        return True # Indicate success

    except FileNotFoundError:
        print("❌ 'kubectl' command not found. Cannot start port-forward.", file=sys.stderr)
        port_forward_process = None
        return False
    except Exception as e:
        print(f"❌ An unexpected error occurred while starting port-forward: {e}", file=sys.stderr)
        # Attempt to clean up if process started but threw later error
        if port_forward_process and port_forward_process.poll() is None:
            port_forward_process.kill()
        port_forward_process = None
        return False

def cleanup(signum=None, frame=None):
    """Cleans up background processes on exit."""
    global port_forward_process
    print("\n🧹 Cleaning up...")
    if port_forward_process and port_forward_process.poll() is None: # Check if it exists and is running
        print(f"   -> Terminating port-forward process (PID: {port_forward_process.pid})...")
        # Try terminate first, then kill
        port_forward_process.terminate() # Send SIGTERM
        try:
            port_forward_process.wait(timeout=3) # Wait briefly for graceful termination
            print("   -> Port-forward terminated gracefully.")
        except subprocess.TimeoutExpired:
            print("   -> Port-forward did not terminate gracefully, sending SIGKILL...")
            port_forward_process.kill() # Force kill
            try:
                 port_forward_process.wait(timeout=2) # Wait for kill
                 print("   -> Port-forward killed.")
            except subprocess.TimeoutExpired:
                 print("   -> Failed to confirm port-forward process termination.")
        except Exception as e:
             print(f"   -> Error during port-forward termination: {e}")
    else:
         print("   -> No active port-forward process found or it already exited.")
    print("👋 Exiting.")
    sys.exit(0)


# --- Main Execution ---

def main():
    # Register cleanup function for graceful exit on Ctrl+C, TERM signal
    signal.signal(signal.SIGINT, cleanup)
    signal.signal(signal.SIGTERM, cleanup)

    print("🚀 Go App Kubernetes Deployer (Network Accessible) 🚀")
    print("-" * 55)

    # 1. Install dependencies
    install_requirements()

    # 2. Ensure Docker is running
    ensure_docker_running()

    # 3. Get user input
    while True:
        folder_path_input = input("📂 Enter the path to the Go application folder: ").strip()
        if not folder_path_input:
            print("🚫 Folder path cannot be empty.", file=sys.stderr)
            continue
        # Try to resolve relative paths (like .)
        folder_name = os.path.abspath(folder_path_input)
        if not os.path.isdir(folder_name):
            print(f"🚫 Folder '{folder_name}' not found or is not a directory.", file=sys.stderr)
        else:
            break # Valid folder found

    port = get_valid_port()
    cluster_name = "go-cluster" # Hardcode kind cluster name for simplicity
    secret_name = "app-secret" # Hardcode secret name

    print(f"\n[*] Configuration Summary:")
    print(f"   -> App Folder:   {folder_name}")
    print(f"   -> App Port:     {port}")
    print(f"   -> Kind Cluster: {cluster_name}")
    print(f"   -> K8s Secret:   {secret_name}")
    print("-" * 55)


    # 4. Prepare files (Dockerfile, k8s manifests, default main.go)
    # Use folder name as base for image and app names, ensure k8s compliance
    base_name = os.path.basename(folder_name).lower().replace("_", "-").replace(" ", "-")
    # Ensure it starts/ends with alphanumeric and contains only alphanumeric/hyphen
    base_name = ''.join(c for c in base_name if c.isalnum() or c == '-')
    base_name = base_name.strip('-')
    if not base_name: base_name = "go-app" # Default if name becomes empty

    image_name = f"{base_name}-image:latest"
    app_name = write_files(folder_name, port, image_name, secret_name)


    # 5. Ensure Kind cluster exists
    if not ensure_kind_cluster(cluster_name):
        sys.exit(1) # Exit if cluster creation failed

    # 6. Create/Update Kubernetes Secret from .env file (before deployment needs it)
    create_k8s_secret_from_env(folder_name, secret_name)

    # 7. Build Docker Image (if needed)
    rebuild = False
    if image_exists_locally(image_name):
         print(f"ℹ️ Image '{image_name}' already exists locally.")
         rebuild_choice = input("   -> Rebuild image anyway? (y/N): ").strip().lower()
         if rebuild_choice == 'y':
            rebuild = True
    if rebuild or not image_exists_locally(image_name):
        build_image(folder_name, image_name)


    # 8. Load Image into Kind
    load_image_to_kind(image_name, cluster_name)


    # 9. Apply Kubernetes Resources
    apply_k8s_resources(folder_name)


    # 10. Wait for Deployment Readiness
    if not wait_for_deployment(app_name):
         print("❌ Deployment failed to become ready. Please check the logs above.", file=sys.stderr)
         print("   -> You might need to manually clean up resources: ")
         print(f"      kubectl delete deployment {app_name}-deployment")
         print(f"      kubectl delete service {app_name}-service")
         # Only suggest deleting secret if .env was used? For simplicity, always suggest.
         print(f"      kubectl delete secret {secret_name} --ignore-not-found")
         sys.exit(1)


    # 11. Start Port Forwarding (with network access enabled)
    service_name = f"{app_name}-service"
    if start_port_forward(service_name, port):
        # 12. Open Browser (only if port-forward started successfully) - Still opens localhost
        local_access_url = f"http://localhost:{port}"
        print(f"[*] Attempting to open application locally in your browser at: {local_access_url}")
        try:
            webbrowser.open(local_access_url)
        except Exception as e:
             print(f"⚠️ Could not automatically open browser: {e}", file=sys.stderr)

        # 13. Final Instructions
        print("\n" + "=" * 60)
        print("✅ Deployment Successful & Port-Forwarding Started!")
        print(f"   Your application '{app_name}' should be accessible:")
        print(f"     -> From THIS computer at:     http://localhost:{port}")
        print(f"     -> From OTHER devices on your network at: http://<your-pc-ip>:{port}")
        print(f"        (Replace <your-pc-ip> with the actual IP address of this computer)")
        print("-" * 60)
        print("   ⚠️ WARNING: You have exposed this service to your local network.")
        print("      Ensure your network is secure and you understand the implications.")
        print("-" * 60)
        print("ℹ️ Port-forwarding is running in the background.")
        print("   Press CTRL+C in this terminal to stop port-forwarding and exit the script.")
        print("=" * 60)


        # Keep the script running so port-forwarding continues
        print("\n⏳ Script is running, keeping port-forward active. Press Ctrl+C to stop.")
        try:
            while True:
                # Check if the background process is still alive
                if port_forward_process and port_forward_process.poll() is not None:
                     print("\n❌ Port-forward process seems to have stopped unexpectedly.", file=sys.stderr)
                     # Try to get output if it stopped
                     try:
                          stdout, stderr = port_forward_process.communicate(timeout=1)
                          if stdout: print(f"   stdout: {stdout.strip()}", file=sys.stderr)
                          if stderr: print(f"   stderr: {stderr.strip()}", file=sys.stderr)
                     except Exception as e:
                          print(f"   (Could not get output from stopped process: {e})")
                     break # Exit the loop
                time.sleep(5) # Check every 5 seconds
        except KeyboardInterrupt:
             cleanup() # Call cleanup on Ctrl+C here as well
        finally:
             # Ensure cleanup runs if the loop breaks for other reasons
             cleanup()

    else:
        # Port forward failed to start
        print("\n" + "=" * 60)
        print("❌ Deployment completed, but FAILED to start port-forwarding.")
        print("   Your application is running in the kind cluster, but is not accessible.")
        print("   Troubleshooting hints from the error messages above might help.")
        print("   Common causes include:")
        print(f"     - Local port {port} already being used by another application.")
        print(f"     - The Kubernetes service 'svc/{service_name}' or its backing pods are not running correctly.")
        print("   You can try running port-forward manually to debug:")
        print(f"     kubectl port-forward svc/{service_name} {port}:{port} --address 0.0.0.0")
        print("=" * 60)
        sys.exit(1) # Exit with error if port-forward failed


if __name__ == "__main__":
    main()