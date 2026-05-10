#include "fs_ops.h"
#include <unistd.h>
#include <sys/wait.h>
#include <stdio.h>
#define NUM_CHILDREN  10
#define NUM_FILES     100
#define FILES_PER_CHILD (NUM_FILES / NUM_CHILDREN)

int main() {
    pid_t pid;

    for (int i = 0; i < NUM_CHILDREN; i++) {
        pid = fork();
        if (pid < 0)
            return 1;

        if (pid == 0) {
            char filepath[64];
            int start = i * FILES_PER_CHILD + 1;
            int end   = start + FILES_PER_CHILD;

            for (int j = start; j < end; j++) {
                // read
                snprintf(filepath, sizeof(filepath), "test_%d.txt", j);
                do_open_read(filepath);

                // write
                snprintf(filepath, sizeof(filepath), "test_%d.txt", j);
                do_open_write(filepath, "stress test content\n");

                // exec
                do_exec("./dummy");
            }

            return 0;
        }
    }

    for (int i = 0; i < NUM_CHILDREN; i++)
        wait(NULL);

    return 0;
}