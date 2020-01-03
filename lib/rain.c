#include "rain.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char** argv) {
    char* in = "foo: bar\n";

    char* out = ToJson(in);

    printf("%s\n", out);

    free(out);
}
