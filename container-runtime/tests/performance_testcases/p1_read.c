#include <stdio.h>
#include <time.h>
#include "fs_ops.h"

#define NUM_FILES 50

int main() {
    char filepath[64];
    int success = 0;
    struct timespec start, end;

    clock_gettime(CLOCK_MONOTONIC, &start);

    for (int i = 1; i <= NUM_FILES; i++) {
        snprintf(filepath, sizeof(filepath), "test_1/test_%d.txt", i);
        if (do_open_read(filepath) == 0)
            success++;
    }

    clock_gettime(CLOCK_MONOTONIC, &end);

    double elapsed = (end.tv_sec - start.tv_sec) +
                     (end.tv_nsec - start.tv_nsec) / 1e9;

    printf("Total time: %.6f seconds\n", elapsed);

    return 0;
}