import importlib

pipeFile = "/tmp/proxy-in.log"
pipeResFile = "/tmp/proxy-out.log"

def sp_exit():
    with open(pipeFile, "w", buffering=1) as pf:
        with open(pipeResFile) as rd:
            pf.write("x\n")
            rd.read(3)

def sp_import_std(lib):
    with open(pipeFile, "w", buffering=1) as pf:
        with open(pipeResFile) as rd:
            pf.write(f"fd{lib}\n")
            rd.read(2)
            pf.write(f"fd{lib}.py\n")
            rd.read(2)
    
    return importlib.import_module(lib)
