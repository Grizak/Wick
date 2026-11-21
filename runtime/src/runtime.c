#include "args/args.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

extern int wick_main_main();

int main(int argc, char *argv[]) {
  args_init(argc, argv);
  int code = wick_main_main();
  return code;
}

void print(const char *message) {
  (void)puts(message);
}

int atoi(const char *str) {
  return (int)strtol(str, NULL, 10);
}

int wick_strlen(const char *str) {
  return strlen(str);
}

char *strmerge(const char *str1, const char *str2) {
  size_t len1 = strlen(str1);
  size_t len2 = strlen(str2);
  char *result = (char *)malloc(len1 + len2 + 1);
  if (!result) {
    return NULL; // Handle memory allocation failure
  }
  memcpy(result, str1, len1);
  memcpy(result + len1, str2, len2);
  result[len1 + len2] = '\0';
  return result;
}

int argc() {
  return args_get_argc();
}

char **argv() {
  return args_get_argv();
}