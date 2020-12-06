import requests
import json
import sys

numDataPoints = sys.argv[1] if len(sys.argv) == 2 else "100"
url = "https://random-word-api.herokuapp.com/word?number="

keys = requests.get(url + str(numDataPoints)).json()
data = requests.get(url + str(numDataPoints)).json()

for w in zip(keys, data):
    requests.put("http://localhost:3000/data", data=json.dumps({"key": w[0], "data": w[1]}))