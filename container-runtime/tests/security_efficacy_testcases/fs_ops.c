#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>
#include <string.h>
#include <errno.h>
#include <sys/wait.h>

#define BUF_SIZE 256

/* ─────────────────────────────────────────
 * do_open: opens a file, triggers FAN_OPEN
 * returns fd on success, -1 on failure
 * ───────────────────────────────────────── */
int do_open(const char *filepath, int flags) {
    int fd = open(filepath, flags);
    if (fd < 0) {
        fprintf(stderr, "[open] Failed to open %s: %s\n",
                filepath, strerror(errno));
        return -1;
    }
    printf("[open] FAN_OPEN  → %s\n", filepath);
    return fd;
}

/* ─────────────────────────────────────────
 * do_read: reads from fd, triggers FAN_ACCESS
 * returns bytes read on success, -1 on failure
 * ───────────────────────────────────────── */
ssize_t do_read(int fd, const char *filepath) {
    char buf[BUF_SIZE];
    ssize_t bytes = read(fd, buf, BUF_SIZE);
    if (bytes < 0) {
        fprintf(stderr, "[read] Failed to read %s: %s\n",
                filepath, strerror(errno));
        return -1;
    }
    printf("[read] FAN_ACCESS → %s (%zd bytes)\n", filepath, bytes);
    return bytes;
}

/* ─────────────────────────────────────────
 * do_write: writes to fd, triggers FAN_MODIFY
 * returns bytes written on success, -1 on failure
 * ───────────────────────────────────────── */
ssize_t do_write(int fd, const char *filepath,
                 const char *data) {
    ssize_t bytes = write(fd, data, strlen(data));
    if (bytes < 0) {
        fprintf(stderr, "[write] Failed to write %s: %s\n",
                filepath, strerror(errno));
        return -1;
    }
    printf("[write] FAN_MODIFY → %s (%zd bytes)\n", filepath, bytes);
    return bytes;
}

/* ─────────────────────────────────────────
 * do_exec: executes a binary, triggers FAN_OPEN_EXEC
 * binary must be a static executable (no shared libs)
 * ───────────────────────────────────────── */
int do_exec(const char *binary) {
    pid_t pid = fork();
    if (pid < 0) {
        fprintf(stderr, "[exec] Fork failed: %s\n", strerror(errno));
        return -1;
    }

    if (pid == 0) {
        // child process
        char *argv[] = { (char *)binary, NULL };
        execv(binary, argv);
        // if execv returns, it failed
        fprintf(stderr, "[exec] execv failed for %s: %s\n",
                binary, strerror(errno));
        exit(1);
    }

    // parent waits for child
    int status;
    waitpid(pid, &status, 0);
    printf("[exec] FAN_OPEN_EXEC → %s\n", binary);
    return WEXITSTATUS(status);
}

int do_open_read(const char *filepath) {
    int fd = open(filepath, O_RDONLY);
    if (fd < 0) {
        fprintf(stderr, "[read] Failed to open %s: %s\n",
                filepath, strerror(errno));
        return -1;
    }
    printf("[open] FAN_OPEN → %s\n", filepath);

    if (do_read(fd, filepath) < 0) {
        close(fd);
        return -1;
    }

    close(fd);
    return 0;
}

// triggers FAN_OPEN + FAN_MODIFY
int do_open_write(const char *filepath, const char *data) {
    int fd = open(filepath, O_WRONLY | O_CREAT | O_TRUNC, 0644);
    if (fd < 0) {
        fprintf(stderr, "[write] Failed to open %s: %s\n",
                filepath, strerror(errno));
        return -1;
    }
    printf("[open] FAN_OPEN → %s\n", filepath);

    if (do_write(fd, filepath, data) < 0) {
        close(fd);
        return -1;
    }

    close(fd);
    return 0;
}