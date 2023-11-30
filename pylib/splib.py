pipeFile = "/tmp/proxy-in.log"
pipeResFile = "/tmp/proxy-out.log"

def sp_exit():
    with open(pipeFile, "w", buffering=1) as pf:
        with open(pipeResFile) as rd:
            pf.write("x\n")
            x = rd.read(3)
            print(x, flush=True)