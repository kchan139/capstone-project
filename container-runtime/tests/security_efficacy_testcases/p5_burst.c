#include "fs_ops.h"
#include <fcntl.h>
#include <unistd.h>

#define NUM_BURST 1000

int main() {
    char buf[256];

    // open once - triggers FAN_OPEN once
    int fd = open("burst_file.txt", O_RDONLY);
    if (fd < 0)
        return 1;

    // read 1000 times on same fd - triggers FAN_ACCESS 1000 times
    // but coalescing will merge them in the queue
    for (int i = 0; i < NUM_BURST; i++) {
        lseek(fd, 0, SEEK_SET); // reset to beginning
        read(fd, buf, sizeof(buf));
    }

    close(fd);
    return 0;
}