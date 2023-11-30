pipeFile = "/tmp/proxy-in.log"

def sp_exit():
    with open(pipeFile, "w") as pf:
        pf.write("x\n")