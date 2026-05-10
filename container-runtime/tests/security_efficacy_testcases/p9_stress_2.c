#include "fs_ops.h"
#include <unistd.h>
#include <sys/wait.h>
#include <stdio.h>

#define NUM_CHILDREN      100
#define NUM_FILES         5000
#define FILES_PER_CHILD   (NUM_FILES / NUM_CHILDREN)

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
                // read 50 files per child from read_ directory
                snprintf(filepath, sizeof(filepath),
                         "test_9/read_%d.txt", j);
                do_open_read(filepath);

                // write 50 files per child to write_ directory
                snprintf(filepath, sizeof(filepath),
                         "test_9/write_%d.txt", j);
                do_open_write(filepath, "stress test content\n");
            }

            // exec p9_helper once per child (handles 100 dummy execs)
            do_exec("./p9_helper");

            return 0;
        }
    }

    for (int i = 0; i < NUM_CHILDREN; i++)
        wait(NULL);

    return 0;
}