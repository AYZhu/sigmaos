import splib
os = splib.sp_import_std("os")
track = splib.sp_import_std("lolthispackageisntreal")

print("hello, world!")
print("preload", os.environ.get("LD_PRELOAD"), flush=True)

splib.sp_exit()