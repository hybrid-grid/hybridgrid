CC ?= gcc
CFLAGS = -Wall -Wextra -O2
SRCS = main.c math_utils.c string_utils.c
OBJS = $(SRCS:.c=.o)
TARGET = testapp

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^

%.o: %.c
	$(CC) $(CFLAGS) -c $<

clean:
	rm -f $(OBJS) $(TARGET)
