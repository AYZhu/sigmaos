#define _GNU_SOURCE
#include <fcntl.h>  /* Definition of AT_* constants */
#include <sys/stat.h>
#include <stdio.h>
#include <stdarg.h>
#include <dlfcn.h>

int fstat(int fd, struct stat *st)
{
    static int (*fstat_func)(int, struct stat*) = NULL;
    fstat_func = (int(*)(int, struct stat*)) dlsym(RTLD_NEXT, "fstat");
    int res = fstat_func(fd, st);
    printf("preloaded fstat\n");
    fflush(stdout);
    return res;
}
int open(const char *filename, int flags, ...)
{

    mode_t mode = 0;

	if (flags & O_CREAT) {
		va_list ap;
		va_start(ap, flags);
		mode = va_arg(ap, mode_t);
		va_end(ap);
	}

    static int (*open_func)(const char*, int, ...) = NULL;
    open_func = (int(*)(const char*, int, ...)) dlsym(RTLD_NEXT, "open");
    int res = open_func(filename, flags, mode);
    printf("preloaded open\n");
    fflush(stdout);
    return res;
}