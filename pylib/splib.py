import importlib
import os

pipeFile = "/tmp/proxy-in.log"
pipeResFile = "/tmp/proxy-out.log"

def sp_exit():
    with open(pipeFile, "w", buffering=1) as pf:
        with open(pipeResFile) as rd:
            pf.write("x\n")
            rd.read(3)

def wait_done(rd):
    x = rd.read(1)
    while x != 'd':
        x = rd.read(1)

def sp_import_std(lib):
    with open(pipeFile, "w", buffering=1) as pf:
        with open(pipeResFile) as rd:
            pf.write(f"pdpylib/Lib/{lib}\n")
            wait_done(rd)
            pf.write(f"pfpylib/Lib/{lib}.py\n")
            wait_done(rd)
    
    importlib.invalidate_caches()
    print(os.listdir("/bin"), flush=True)
    try:
        return importlib.import_module(lib)
    except ModuleNotFoundError as e:
        if e.name != lib:
            sp_import_std(e.name)
            return sp_import_std(lib)
        else: 
            raise e
