import os
import subprocess
import sys
import shutil
import platform
import socket

DOCKERFILE_TEMPLATE = '''
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE {port}
CMD ["./main"]
'''

DEPLOYMENT_YAML_TEMPLATE = '''
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {name}-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {name}
  template:
    metadata:
      labels:
        app: {name}
    spec:
      containers:
      - name: {name}-container
        image: {name}-image
        ports:
        - containerPort: {port}
'''

SERVICE_YAML_TEMPLATE = '''
apiVersion: v1
kind: Service
metadata:
  name: {name}-service
spec:
  type: NodePort
  selector:
    app: {name}
  ports:
    - port: 80
      targetPort: {port}
      nodePort: 30001
'''

def check_and_install_tool(tool_name, install_command):
    if shutil.which(tool_name) is None:
        print(f"🚫 {tool_name} tidak ditemukan. Menginstal...")
        subprocess.run(install_command, shell=True, check=True)
    else:
        print(f"[*] {tool_name} sudah terinstal.")

def is_port_in_use(port):
    """Cek apakah port sedang digunakan di localhost."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        return sock.connect_ex(("127.0.0.1", port)) == 0

def get_valid_port():
    while True:
        port_input = input("🔌 Masukkan port aplikasi (contoh: 8080): ").strip()
        if not port_input.isdigit():
            print("🚫 Port harus berupa angka.")
            continue

        port = int(port_input)
        if is_port_in_use(port):
            print(f"❌ Port {port} sedang digunakan. Coba port lain.")
        else:
            return port

def write_files(folder_name, port):
    go_file_path = os.path.join(folder_name, "main.go")
    if not os.path.exists(go_file_path):
        print(f"🚫 File '{go_file_path}' tidak ditemukan.")
        sys.exit(1)

    name = os.path.basename(folder_name)

    with open(os.path.join(folder_name, "Dockerfile"), "w") as f:
        f.write(DOCKERFILE_TEMPLATE.format(port=port))

    with open(os.path.join(folder_name, "deployment.yaml"), "w") as f:
        f.write(DEPLOYMENT_YAML_TEMPLATE.format(name=name, port=port))

    with open(os.path.join(folder_name, "service.yaml"), "w") as f:
        f.write(SERVICE_YAML_TEMPLATE.format(name=name, port=port))

    return name

def build_image(folder_name, image_name):
    print(f"[*] Building image {image_name}...")
    subprocess.run(["docker", "build", "-t", image_name, "."], cwd=folder_name, check=True)

def ensure_kind_cluster():
    print("[*] Memastikan kind cluster tersedia...")
    clusters = subprocess.check_output(["kind", "get", "clusters"]).decode()
    if "go-cluster" not in clusters:
        subprocess.run(["kind", "create", "cluster", "--name", "go-cluster"], check=True)

def load_image_to_kind(image_name):
    print(f"[*] Loading image {image_name} ke kind...")
    subprocess.run(["kind", "load", "docker-image", image_name, "--name", "go-cluster"], check=True)

def apply_k8s_resources(folder_name):
    print("[*] Men-deploy ke Kubernetes...")
    subprocess.run(["kubectl", "apply", "-f", "deployment.yaml"], cwd=folder_name, check=True)
    subprocess.run(["kubectl", "apply", "-f", "service.yaml"], cwd=folder_name, check=True)

def main():
    if platform.system() == "Linux":
        check_and_install_tool("docker", "sudo apt-get install -y docker.io")
        check_and_install_tool("kubectl", "sudo apt-get install -y kubectl")
        check_and_install_tool("kind", "sudo curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64 && sudo chmod +x /usr/local/bin/kind")
    elif platform.system() == "Darwin":
        check_and_install_tool("docker", "brew install --cask docker")
        check_and_install_tool("kubectl", "brew install kubectl")
        check_and_install_tool("kind", "brew install kind")
    else:
        print("🚫 OS belum didukung.")
        sys.exit(1)

    folder_name = input("📁 Masukkan nama folder aplikasi Go: ").strip()
    if not os.path.isdir(folder_name):
        print(f"🚫 Folder '{folder_name}' tidak ditemukan.")
        sys.exit(1)

    port = get_valid_port()

    print(f"🚀 Deploying aplikasi dari folder '{folder_name}' ke Kubernetes dengan port {port}...")

    name = write_files(folder_name, port)
    ensure_kind_cluster()
    build_image(folder_name, f"{name}-image")
    load_image_to_kind(f"{name}-image")
    apply_k8s_resources(folder_name)

    print("✅ Deployment selesai.")
    print(f"👉 Akses aplikasi dengan: kubectl port-forward svc/{name}-service {port}:80")
    print(f"🌐 Lalu buka: http://localhost:{port}")

if __name__ == "__main__":
    if not shutil.which("docker") or not shutil.which("kubectl") or not shutil.which("kind"):
        print("🚫 Pastikan Docker, kubectl, dan kind sudah terinstal.")
        sys.exit(1)
    main()
