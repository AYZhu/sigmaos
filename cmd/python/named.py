import splib
import sys

print(sys.argv, flush=True)
if len(sys.argv) < 2:
    raise ValueError("expected at least two arguments")

prog = ""

with splib.NamedReader(sys.argv[1], False) as r:
    prog = r.fd.read()

exec(prog)