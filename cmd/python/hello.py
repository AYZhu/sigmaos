import splib
os = splib.sp_import_std("os")

print("hello, world!")
print("preload", os.environ.get("LD_PRELOAD"), flush=True)

datetime = splib.sp_import_std("datetime")

splib.sp_exit()