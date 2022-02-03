import requests

r = requests.post("http://127.0.0.1:8000/class/new", data={"teacher_id": 1, "name": "9.b"})

print(r.text, r.status_code)
