#include <stdio.h>
#include <stdlib.h>
#include "math_utils.h"
#include "string_utils.h"

int main(void) {
    printf("=== Math Utils Test ===\n");
    printf("add(5, 3) = %d\n", add(5, 3));
    printf("multiply(4, 7) = %d\n", multiply(4, 7));
    printf("factorial(5) = %d\n", factorial(5));

    printf("\n=== String Utils Test ===\n");
    printf("string_length(\"hello\") = %zu\n", string_length("hello"));

    char *rev = string_reverse("hello");
    if (rev) {
        printf("string_reverse(\"hello\") = \"%s\"\n", rev);
        free(rev);
    }

    char *concat = string_concat("hello", "world");
    if (concat) {
        printf("string_concat(\"hello\", \"world\") = \"%s\"\n", concat);
        free(concat);
    }

    printf("\n=== All tests completed ===\n");
    return 0;
}
