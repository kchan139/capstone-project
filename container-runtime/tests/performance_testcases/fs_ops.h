#ifndef FS_OPS_H
#define FS_OPS_H

#include <sys/types.h>

ssize_t do_read(int fd, const char *filepath);
ssize_t do_write(int fd, const char *filepath, const char *data);
int     do_open_read(const char *filepath);
int     do_open_write(const char *filepath, const char *data);
int     do_exec(const char *binary);

#endif