import requests

r1 = requests.post("http://127.0.0.1:8000/user/new", data={"email": "test1@meetplan.si", "pass": "test", "name": "Test 1"})
r2 = requests.post("http://127.0.0.1:8000/user/new", data={"email": "test2@meetplan.si", "pass": "test", "name": "Test 2"})

print(r1.text, r1.status_code)
print(r2.text, r2.status_code)
