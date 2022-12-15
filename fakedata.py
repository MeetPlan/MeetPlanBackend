# This is a Python script to fill MeetPlan with fake data (users, subjects and classes). Highly recommended for testing purposes.
# Requirements:
# - httpx

import httpx
import asyncio
import time

# Default debug config of MeetPlan
URL = "http://localhost:8000/"
EMAIL_END = "meetplan.si"
SCHOOL_YEAR = "2022/2023"

class User:
    def __init__(self, name, password, email):
        self.name = name
        self.password = password
        self.email = email

class Class:
    def __init__(self, name, students, teacher):
        self.name = name
        self.students = students
        self.teacher = teacher

class Subject:
    def __init__(self, name, long, students, teacher, is_graded=True):
        self.name = name
        self.long = long
        self.is_graded = is_graded
        self.students = students
        self.teacher = teacher

users = [
    User("Administrator", "admin", f"admin@{EMAIL_END}"),
    User("Ravnatelj", "principal", f"principal@{EMAIL_END}"),
    User("Pomočnik ravnatelja", "principal assistant", f"principalassistant@{EMAIL_END}"),
    User("Šolski psiholog", "school psychologist", f"schoolpsychologist@{EMAIL_END}"),
    User("Organizator šolske prehrane", "food", f"food@{EMAIL_END}"),
    User("Učitelj jezikov 1", "teacher", f"lang1@{EMAIL_END}"),
    User("Učitelj jezikov 2", "teacher", f"lang2@{EMAIL_END}"),
    User("Učitelj jezikov 3", "teacher", f"lang3@{EMAIL_END}"),
    User("Učitelj matematike", "teacher", f"math@{EMAIL_END}"),
    User("Učitelj biologije", "teacher", f"biology@{EMAIL_END}"),
    User("Učitelj kemije", "teacher", f"chemistry@{EMAIL_END}"),
    User("Učitelj fizike", "teacher", f"physics@{EMAIL_END}"),
    User("Učitelj naravoslovnih predmetov", "teacher", f"naturalsci@{EMAIL_END}"),
    User("Učitelj geografije", "teacher", f"geography@{EMAIL_END}"),
    User("Učitelj zgodovine", "teacher", f"history@{EMAIL_END}"),
    User("Učitelj likovne umetnosti", "teacher", f"art@{EMAIL_END}"),
    User("Učitelj glasbe", "teacher", f"music@{EMAIL_END}"),
    User("Učitelj športa", "teacher", f"sports@{EMAIL_END}"),
    User("Učenec 1", "student", f"student1@{EMAIL_END}"),
    User("Učenec 2", "student", f"student2@{EMAIL_END}"),
    User("Učenec 3", "student", f"student3@{EMAIL_END}"),
    User("Učenec 4", "student", f"student4@{EMAIL_END}"),
    User("Učenec 5", "student", f"student5@{EMAIL_END}"),
    User("Učenec 6", "student", f"student6@{EMAIL_END}"),
    User("Učenec 7", "student", f"student7@{EMAIL_END}"),
    User("Učenec 8", "student", f"student8@{EMAIL_END}"),
    User("Učenec 9", "student", f"student9@{EMAIL_END}"),
    User("Starš 1", "parent", f"parent1@{EMAIL_END}"),
    User("Starš 2", "parent", f"parent2@{EMAIL_END}"),
    User("Starš 3", "parent", f"parent3@{EMAIL_END}"),
    User("Starš 4", "parent", f"parent4@{EMAIL_END}"),
]

classes = [
    Class("8.a", [f"student1@{EMAIL_END}", f"student2@{EMAIL_END}"], f"lang1@{EMAIL_END}"),
    Class("8.b", [f"student3@{EMAIL_END}", f"student4@{EMAIL_END}"], f"biology@{EMAIL_END}"),
    Class("9.a", [f"student5@{EMAIL_END}", f"student6@{EMAIL_END}"], f"chemistry@{EMAIL_END}"),
    Class("9.b", [f"student7@{EMAIL_END}", f"student8@{EMAIL_END}", f"student9@{EMAIL_END}"], f"art@{EMAIL_END}"),
]

subjects = [
    # 8.a
    Subject("RU8a", "razredna ura", "8.a", f"lang1@{EMAIL_END}", False),
    Subject("LUM8a", "likovna umetnost", "8.a", f"art@{EMAIL_END}"),
    Subject("GUM8a", "glasbena umetnost", "8.a", f"music@{EMAIL_END}"),
    Subject("GEO8a", "geografija", "8.a", f"geography@{EMAIL_END}"),
    Subject("ZGO8a", "zgodovina", "8.a", f"history@{EMAIL_END}"),
    Subject("DKE8a", "domovinska in državljanska kultura ter etika", "8.a", f"geography@{EMAIL_END}"),
    Subject("FIZ8a", "fizika", "8.a", f"physics@{EMAIL_END}"),
    Subject("KEM8a", "kemija", "8.a", f"chemistry@{EMAIL_END}"),
    Subject("BIO8a", "biologija", "8.a", f"biology@{EMAIL_END}"),
    Subject("TIT8a", "tehnika in tehnologija", "8.a", f"naturalsci@{EMAIL_END}"),
    Subject("ŠPO8a", "šport", "8.a", f"sports@{EMAIL_END}"),

    # 8.b
    Subject("RU8b", "razredna ura", "8.b", f"biology@{EMAIL_END}", False),
    Subject("LUM8b", "likovna umetnost", "8.b", f"art@{EMAIL_END}"),
    Subject("GUM8b", "glasbena umetnost", "8.b", f"music@{EMAIL_END}"),
    Subject("GEO8b", "geografija", "8.b", f"geography@{EMAIL_END}"),
    Subject("ZGO8b", "zgodovina", "8.b", f"history@{EMAIL_END}"),
    Subject("DKE8b", "domovinska in državljanska kultura ter etika", "8.b", f"history@{EMAIL_END}"),
    Subject("FIZ8b", "fizika", "8.b", f"physics@{EMAIL_END}"),
    Subject("KEM8b", "kemija", "8.b", f"chemistry@{EMAIL_END}"),
    Subject("BIO8b", "biologija", "8.b", f"biology@{EMAIL_END}"),
    Subject("TIT8b", "tehnika in tehnologija", "8.b", f"naturalsci@{EMAIL_END}"),
    Subject("ŠPO8b", "šport", "8.b", f"sports@{EMAIL_END}"),

    # Mešane skupine 8. razreda
    Subject("MAT8a", "matematika", [f"student1@{EMAIL_END}", f"student3@{EMAIL_END}"], f"math@{EMAIL_END}"),
    Subject("MAT8b", "matematika", [f"student2@{EMAIL_END}", f"student4@{EMAIL_END}"], f"naturalsci@{EMAIL_END}"),
    Subject("SLJ8a", "slovenščina", [f"student1@{EMAIL_END}", f"student3@{EMAIL_END}"], f"lang2@{EMAIL_END}"),
    Subject("SLJ8b", "slovenščina", [f"student2@{EMAIL_END}", f"student4@{EMAIL_END}"], f"lang3@{EMAIL_END}"),
    Subject("TJA8a", "angleščina", [f"student1@{EMAIL_END}", f"student3@{EMAIL_END}"], f"lang1@{EMAIL_END}"),
    Subject("TJA8b", "angleščina", [f"student2@{EMAIL_END}", f"student4@{EMAIL_END}"], f"lang3@{EMAIL_END}"),

    # 9.a
    Subject("RU9a", "razredna ura", "9.a", f"chemistry@{EMAIL_END}", False),
    Subject("LUM9a", "likovna umetnost", "9.a", f"art@{EMAIL_END}"),
    Subject("GUM9a", "glasbena umetnost", "9.a", f"music@{EMAIL_END}"),
    Subject("GEO9a", "geografija", "9.a", f"geography@{EMAIL_END}"),
    Subject("ZGO9a", "zgodovina", "9.a", f"history@{EMAIL_END}"),
    Subject("FIZ9a", "fizika", "9.a", f"physics@{EMAIL_END}"),
    Subject("KEM9a", "kemija", "9.a", f"chemistry@{EMAIL_END}"),
    Subject("BIO9a", "biologija", "9.a", f"biology@{EMAIL_END}"),
    Subject("ŠPO9a", "šport", "9.a", f"sports@{EMAIL_END}"),

    # 9.b
    Subject("RU9b", "razredna ura", "9.b", f"art@{EMAIL_END}", False),
    Subject("LUM9b", "likovna umetnost", "9.b", f"art@{EMAIL_END}"),
    Subject("GUM9b", "glasbena umetnost", "9.b", f"music@{EMAIL_END}"),
    Subject("GEO9b", "geografija", "9.b", f"geography@{EMAIL_END}"),
    Subject("ZGO9b", "zgodovina", "9.b", f"history@{EMAIL_END}"),
    Subject("FIZ9b", "fizika", "9.b", f"physics@{EMAIL_END}"),
    Subject("KEM9b", "kemija", "9.b", f"chemistry@{EMAIL_END}"),
    Subject("BIO9b", "biologija", "9.b", f"biology@{EMAIL_END}"),
    Subject("ŠPO9b", "šport", "9.b", f"sports@{EMAIL_END}"),

    # Mešane skupine 9. razreda
    Subject("MAT9a", "matematika", [f"student5@{EMAIL_END}", f"student9@{EMAIL_END}"], f"math@{EMAIL_END}"),
    Subject("MAT9b", "matematika", [f"student6@{EMAIL_END}", f"student7@{EMAIL_END}", f"student8@{EMAIL_END}"], f"naturalsci@{EMAIL_END}"),
    Subject("SLJ9a", "slovenščina", [f"student5@{EMAIL_END}", f"student9@{EMAIL_END}"], f"lang2@{EMAIL_END}"),
    Subject("SLJ9b", "slovenščina", [f"student6@{EMAIL_END}", f"student7@{EMAIL_END}", f"student8@{EMAIL_END}"], f"lang3@{EMAIL_END}"),
    Subject("TJA9a", "angleščina", [f"student5@{EMAIL_END}", f"student9@{EMAIL_END}"], f"lang1@{EMAIL_END}"),
    Subject("TJA9b", "angleščina", [f"student6@{EMAIL_END}", f"student7@{EMAIL_END}", f"student8@{EMAIL_END}"], f"lang3@{EMAIL_END}"),

    # Neobvezni izbirni predmeti – mešano
    Subject("NEM8", "nemščina", [f"student1@{EMAIL_END}", f"student4@{EMAIL_END}"], f"lang1@{EMAIL_END}"),
    Subject("NEM9", "nemščina", [f"student5@{EMAIL_END}", f"student6@{EMAIL_END}", f"student8@{EMAIL_END}", f"student9@{EMAIL_END}"], f"lang1@{EMAIL_END}"),
    Subject("MME", "multimedija", [f"student1@{EMAIL_END}", f"student2@{EMAIL_END}", f"student3@{EMAIL_END}"], f"physics@{EMAIL_END}"),
    Subject("ROM", "računalniška omrežja", [f"student5@{EMAIL_END}", f"student6@{EMAIL_END}", f"student7@{EMAIL_END}", f"student9@{EMAIL_END}"], f"naturalsci@{EMAIL_END}"),
]

async def main():
    client = httpx.AsyncClient()
    tstart = time.time()

    for user in users:
        r = await client.post(f"{URL}user/new", data={"email": user.email, "pass": user.password, "name": user.name})
        if r.status_code == 201:
            print(f"[SUCCESS] User {user.name} has been created successfully")
        else:
            print(f"[FAIL] User {user.name} creation failed {r.text}")

    print(f"[DONE] User creation has completed in {time.time()-tstart} seconds.")

    # From now on, we manage all our users as administrators
    r = await client.post(f"{URL}user/login", data={"email": users[0].email, "pass": users[0].password})
    headers = {"Authorization": f"Bearer {await r.json()['data']['token']}"}
    client.headers = headers

    # Map all User IDs to our users
    users_sys = await client.get(f"{URL}users/get")
    for i in (await users_sys.json())["data"]:
        for n in range(len(users)):
            if users[n].email == i["Email"]:
                users[n].id = i["ID"]

    ts = time.time()
    for user in users:
        r = await client.post(f"{URL}user/role/update/{user.id}", data={"role": user.password}) # Conveniently, our passwords are same as roles
        if r.status_code == 200:
            print(f"[SUCCESS] User {user.name}'s role has been successfully changed to {user.password}")
        else:
            print(f"[FAIL] User {user.name}'s role hasn't been successfully changed to {user.password}")
    print(f"[DONE] User role changing has completed in {time.time()-ts} seconds.")

    ts = time.time()
    for i, class_ in enumerate(classes):
        teacher_id = ""
        for user in users:
            if user.password == "teacher" and user.email == class_.teacher:
                teacher_id = user.id
                break
        if teacher_id == "":
            print(f"[FAIL] Teacher {class_.teacher} couldn't be assigned to the class")
        await client.post(f"{URL}class/new", data={"teacher_id": teacher_id, "name": class_.name})

    class_sys = await client.get(f"{URL}classes/get")
    for i in (await class_sys.json())["data"]:
        for n in range(len(classes)):
            if classes[n].name == i["Name"]:
                classes[n].id = i["ID"]

    for i, class_ in enumerate(classes):
        print(f"[INFO] Adding users to class {class_.id} ({class_.name})")
        for class_user in class_.students:
            for user in users:
                if user.email != class_user:
                    continue
                await client.post(f"{URL}class/get/{class_.id}/add_user/{user.id}")
    print(f"[DONE] Class creation has completed in {time.time()-ts} seconds.")

    ts = time.time()
    for i, subject in enumerate(subjects):
        teacher_id = ""
        for user in users:
            if user.password == "teacher" and user.email == subject.teacher:
                teacher_id = user.id
                break
        if teacher_id == "":
            print(f"[FAIL] Teacher {subject.teacher} couldn't be assigned to the class")
        r = await client.post(f"{URL}subject/new", data={
            "teacher_id": teacher_id,
            "name": subject.name,
            "long_name": subject.long,
            "class_id": subject.students if type(subject.students) is str else "",
            "realization": 160,
            "is_graded": subject.is_graded,
            "location": 50,
        })

    subjects_sys = await client.get(f"{URL}subjects/get")
    for i in (await subjects_sys.json())["data"]:
        for n in range(len(subjects)):
            if subjects[n].name == i["Name"]:
                subjects[n].id = i["ID"]

    for i, subject in enumerate(subjects):
        if type(subject.students) is str:
            continue
        print(f"[INFO] Adding users to subject {subject.id} ({subject.name})")
        for subject_user in subject.students:
            for user in users:
                if user.email != subject_user:
                    continue
                await client.post(f"{URL}subject/get/{subject.id}/add_user/{user.id}")
    print(f"[DONE] Subject creation has completed in {time.time()-ts} seconds.")

asyncio.run(main())
