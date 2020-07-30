import csv
import uuid
import sys

NUMBER_OF_USERS = int(sys.argv[1])

with open('users.csv', 'w', newline = '') as file:
    fieldnames = ['firstName', 'lastName', 'login', 'email']
    writer = csv.DictWriter(file, fieldnames=fieldnames, delimiter=',')
    writer.writeheader()
    for i in range(0,NUMBER_OF_USERS):
        uid = uuid.uuid1()
        uid = str(uid).replace("-",".")
        writer.writerow({'firstName' : "first_" + uid, 'lastName' : "last_" + uid, 'login' : uid + '@mydomain.com', 'email' : uid + '@mydomain.com'})
