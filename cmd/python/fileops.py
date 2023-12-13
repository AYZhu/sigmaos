import splib

with splib.NamedReader("name/a.txt", False) as r:
    print(r.fd.read())

with splib.NamedReader("name/a.txt", True) as r:
    r.fd.write("goodbye, test!")

splib.sp_exit()