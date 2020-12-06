import subprocess
import requests
import sys

numCANs = int(sys.argv[1]) if len(sys.argv) >= 2 else 10
port = int(sys.argv[2]) if len(sys.argv) >= 3 else 3001

for c in range(numCANs):
    key = requests.get("https://random-word-api.herokuapp.com/word?number=1").json()[0]
    subprocess.Popen(["go", "run", "../can.go", "-p", str(port), "-join", "localhost:3000", "-key", key])
    port += 1