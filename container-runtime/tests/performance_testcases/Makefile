CC     = gcc
CFLAGS = -Wall -Wextra

p1: p1_read.c fs_ops.c
	$(CC) $(CFLAGS) -o p1 p1_read.c fs_ops.c
p2: p2_write.c fs_ops.c
	$(CC) $(CFLAGS) -o p2 p2_write.c fs_ops.c
p3: p3_exec.c fs_ops.c
	$(CC) $(CFLAGS) -o p3 p3_exec.c fs_ops.c

p4: p4_mix.c fs_ops.c
	$(CC) $(CFLAGS) -o p4 p4_mix.c fs_ops.c

p5: p5_burst.c fs_ops.c
	$(CC) $(CFLAGS) -o p5 p5_burst.c fs_ops.c
p8: p8_stress.c fs_ops.c
	$(CC) $(CFLAGS) -o p8 p8_stress.c fs_ops.c
p9: p9_stress_2.c fs_ops.c
	$(CC) $(CFLAGS) -o p9 p9_stress_2.c fs_ops.c
	$(CC) $(CFLAGS) -o p9_helper p9_helper.c fs_ops.c