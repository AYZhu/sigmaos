import os
import splib

print("hello, world!")
print("preload", os.environ.get("LD_PRELOAD"))

splib.sp_exit()