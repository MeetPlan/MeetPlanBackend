import requests

r = requests.get("http://127.0.0.1:8000/class/get/0/self_testing")

print(r.text, r.status_code)
