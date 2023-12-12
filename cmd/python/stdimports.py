import splib
random = splib.sp_import_std("random")

# currently, skips the following imports
# * asyncio - broken because of `coroutines` for some reason (NameError - should be resolvable?) will need to strace this
# * msilib - windows only
# * ssl - library
# * bz2 - library
# * ctypes - library
# * lzma - library
# * curses - library
# * turtle, tkinter - library
# * crypt
# * pydoc, zoneinfo, ensurepip, cgitb - depends on sysconfig data that is missing, somehow?????????????????????
# * gzip - missing zlib (library)
# * uuid - depends on os.uname(), blocked by seccomp
# * sqlite3 - missing _sqlite3 (library?)


names = ["collections", "concurrent", "dbm", "distutils", "email", "encodings", "html", "http", "idlelib", "json", "lib2to3", "logging", "multiprocessing", "pydoc_data", "re", "test", "tomllib", "unittest", "urllib", "venv", "wsgiref", "xml", "xmlrpc", "sysconfig", "abc", "aifc", "antigravity", "argparse", "ast", "asynchat", "asyncore", "base64", "bdb", "bisect", "calendar", "cgi", "chunk", "cmd", "code", "codecs", "codeop", "colorsys", "compileall", "configparser", "contextlib", "contextvars", "copy", "copyreg", "cProfile", "csv", "dataclasses", "datetime", "decimal", "difflib", "dis", "doctest", "enum", "filecmp", "fileinput", "fnmatch", "fractions", "ftplib", "functools", "genericpath", "getopt", "getpass", "gettext", "glob", "graphlib", "hashlib", "heapq", "hmac", "imaplib", "imghdr", "imp", "inspect", "io", "ipaddress", "keyword", "linecache", "locale", "mailbox", "mailcap", "mimetypes", "modulefinder", "netrc", "nntplib", "ntpath", "nturl2path", "numbers", "opcode", "operator", "optparse", "os", "pathlib", "pdb", "pickle", "pickletools", "pipes", "pkgutil", "platform", "plistlib", "poplib", "posixpath", "pprint", "profile", "pstats", "pty", "py_compile", "pyclbr", "queue", "quopri", "random", "reprlib", "rlcompleter", "runpy", "sched", "secrets", "selectors", "shelve", "shlex", "shutil", "signal", "site", "smtpd", "smtplib", "sndhdr", "socket", "socketserver", "sre_compile", "sre_constants", "sre_parse", "stat", "statistics", "string", "stringprep", "struct", "subprocess", "sunau", "symtable", "tabnanny", "tarfile", "telnetlib", "tempfile", "textwrap", "this", "threading", "timeit", "token", "tokenize", "trace", "traceback", "tracemalloc", "tty", "types", "typing", "uu", "warnings", "wave", "weakref", "webbrowser", "xdrlib", "zipapp", "zipfile", "zipimport" ]
random.shuffle(names)

print(len(names), flush=True)
print(names, flush=True)

for name in names:
    splib.sp_import_std(name)

print("hello, world!")

splib.sp_exit()