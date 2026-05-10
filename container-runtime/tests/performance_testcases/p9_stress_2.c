#include "fs_ops.h"
#include <unistd.h>
#include <sys/wait.h>
#include <stdio.h>
#include <time.h>

#define NUM_CHILDREN      100
#define NUM_FILES         5000
#define FILES_PER_CHILD   (NUM_FILES / NUM_CHILDREN)

int main() {
    pid_t pid;
    struct timespec start, end;

    clock_gettime(CLOCK_MONOTONIC, &start);

    for (int i = 0; i < NUM_CHILDREN; i++) {
        pid = fork();
        if (pid < 0)
            return 1;

        if (pid == 0) {
            char filepath[64];
            int start_idx = i * FILES_PER_CHILD + 1;
            int end_idx   = start_idx + FILES_PER_CHILD;

            for (int j = start_idx; j < end_idx; j++) {
                snprintf(filepath, sizeof(filepath),
                         "test_9/read_%d.txt", j);
                do_open_read(filepath);

                snprintf(filepath, sizeof(filepath),
                         "test_9/write_%d.txt", j);
                do_open_write(filepath, "stress test content\n");
            }

            do_exec("./p9_helper");

            return 0;
        }
    }

    for (int i = 0; i < NUM_CHILDREN; i++)
        wait(NULL);

    clock_gettime(CLOCK_MONOTONIC, &end);

    double elapsed = (end.tv_sec - start.tv_sec) +
                     (end.tv_nsec - start.tv_nsec) / 1e9;

    printf("[P9] Done.\n");
    printf("Total time: %.6f seconds\n", elapsed);

    return 0;
}