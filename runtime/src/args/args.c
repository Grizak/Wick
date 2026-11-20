int args_argc;
char **args_argv;

void args_init(int argc, char *argv[]) {
  args_argc = argc;
  args_argv = argv;
};

int args_get_argc() {
  return args_argc;
}

char **args_get_argv() {
  return args_argv;
}