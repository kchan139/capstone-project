#include "fs_ops.h"
#include <fcntl.h>
#include <unistd.h>
#include <stdio.h>
#include <time.h>

#define NUM_BURST 1000

int main() {
    char buf[256];
    struct timespec start, end;

    int fd = open("test_5/burst_file.txt", O_RDONLY);
    if (fd < 0)
        return 1;

    clock_gettime(CLOCK_MONOTONIC, &start);

    for (int i = 0; i < NUM_BURST; i++) {
        lseek(fd, 0, SEEK_SET);
        read(fd, buf, sizeof(buf));
    }

    clock_gettime(CLOCK_MONOTONIC, &end);

    close(fd);

    double elapsed = (end.tv_sec - start.tv_sec) +
                     (end.tv_nsec - start.tv_nsec) / 1e9;

    printf("[P5] Done. %d reads on same fd.\n", NUM_BURST);
    printf("Total time: %.6f seconds\n", elapsed);

    return 0;
}