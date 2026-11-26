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

#if defined(_MSC_VER)
#pragma comment(linker, "/alternatename:wick_print_impl=wick_print.1")
#else
void wick_print_impl(const char *message) __asm__("wick_print.1");
#endif

void wick_print_impl(const char *message) {
  (void)puts(message);
}

int wick_atoi(const char *str) {
  return (int)strtol(str, NULL, 10);
}

int wick_strlen(const char *str) {
  return strlen(str);
}

char *wick_strmerge(const char *str1, const char *str2) {
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

int wick_argc() {
  return args_get_argc();
}

char **wick_argv() {
  return args_get_argv();
}
