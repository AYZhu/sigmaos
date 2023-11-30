#define _GNU_SOURCE
#include <sys/stat.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <dlfcn.h>
#include <dirent.h>
#include <string.h>
#include <unistd.h>

const char* pipeResFile = "/tmp/proxy-out.log";
const char* pipeFile = "/tmp/proxy-in.log";

const char* get_path(const char *filename)
{
    // funct();
    static int (*open_func)(const char*, int, mode_t) = NULL;
    open_func = (int(*)(const char*, int, mode_t)) dlsym(RTLD_NEXT, "open");

    const char* prefix = "/~~";
    int i = 0;
    while(filename[i] != 0 && i < 3) {
        if (filename[i] != prefix[i]) {
            return filename;
        }
        i++;
    }
    
    if (i < 3) return filename;

    printf("opening read pipe\n");
    fflush(stdout);
    int fd = open_func(pipeFile, 1, 0666); // O_WRONLY
    printf("opening write pipe %s %li\n", filename, strlen(filename));
    fflush(stdout);
    int fdRes = open_func(pipeResFile, 0, 0666); // O_RDONLY 
    char* x = malloc(512 * sizeof(char));
    char x2[3];
    printf("opened pipes\n");
    fflush(stdout);

    write(fd, "l", 1);
    write(fd, &(filename[3]), strlen(filename) - 3);
    write(fd, "\n", 1);
    read(fdRes, x2, 2);
    x2[2] = '\0';
    printf("got back %s\n", x2);

    sprintf(x, "%s%s", "/bin2/pylib/Lib", &(filename[3]));
    // printf("redirected file at %s to %s\n", filename, x);
    return x;
}

int stat(const char *path, struct stat *buf) {
    static int (*fstat_func)(const char*, struct stat*) = NULL;
    fstat_func = (int(*)(const char*, struct stat*)) dlsym(RTLD_NEXT, "stat");
    int res = fstat_func(get_path(path), buf);
    printf("preloaded stat\n");
    fflush(stdout);
    return res;
}

int fstat(int fd, struct stat *st)
{
    static int (*fstat_func)(int, struct stat*) = NULL;
    fstat_func = (int(*)(int, struct stat*)) dlsym(RTLD_NEXT, "fstat");
    int res = fstat_func(fd, st);
    printf("preloaded fstat\n");
    fflush(stdout);
    return res;
}
/*
int open(const char *filename, int flags)
{


    static int (*open_func)(const char*, int) = NULL;
    open_func = (int(*)(const char*, int)) dlsym(RTLD_NEXT, "open");
    int res = open_func(get_path(filename), flags);
    printf("preloaded open\n");
    fflush(stdout);
    return res;
}*/

int open(const char *filename, int flags, mode_t mode)
{
    static int (*open_func)(const char*, int, mode_t) = NULL;
    open_func = (int(*)(const char*, int, mode_t)) dlsym(RTLD_NEXT, "open");
    int res = open_func(get_path(filename), flags, mode);
    printf("preloaded open\n");
    fflush(stdout);
    return res;
}

FILE * fopen( const char * filename,
              const char * mode )
{
    static FILE * (*fopen_func)(const char*, const char*) = NULL;
    fopen_func = (FILE* (*)(const char*, const char*)) dlsym(RTLD_NEXT, "fopen");
    FILE * res = fopen_func(get_path(filename), mode);
    printf("preloaded fopen\n");
    fflush(stdout);
    return res;
}
/**
int openat(int dirfd, const char *pathname, int flags)
{
    static int (*open_func)(int, const char*, int) = NULL;
    open_func = (int(*)(int, const char*, int)) dlsym(RTLD_NEXT, "openat");
    int res = open_func(dirfd, get_path(pathname), flags);
    printf("preloaded openat\n");
    fflush(stdout);
    return res;
}
*/
int openat(int dirfd, const char *pathname, int flags, mode_t mode)
{
    static int (*open_func)(int, const char*, int, mode_t) = NULL;
    open_func = (int(*)(int, const char*, int, mode_t)) dlsym(RTLD_NEXT, "openat");
    int res = open_func(dirfd, get_path(pathname), flags, mode);
    printf("preloaded openat\n");
    fflush(stdout);
    return res;
}
DIR * opendir(const char* name)
{
    static DIR * (*opendir_func)(const char*) = NULL;
    opendir_func = (DIR*(*)(const char*)) dlsym(RTLD_NEXT, "opendir");
    DIR* res = opendir_func(get_path(name));
    printf("preloaded opendir\n");
    fflush(stdout);
    return res;
}