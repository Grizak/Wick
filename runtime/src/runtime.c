#include "args/args.h"
#include <stdio.h>

extern int wick_main_main();

int main(int argc, char *argv[]) {
  args_init(argc, argv);
  int code = wick_main_main();
  return code;
}

void print(const char *message) {
  puts(message);
}