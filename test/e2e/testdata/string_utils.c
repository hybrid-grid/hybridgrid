#include "string_utils.h"
#include <string.h>
#include <stdlib.h>

size_t string_length(const char *str) {
    if (str == NULL) return 0;
    return strlen(str);
}

char *string_reverse(const char *str) {
    if (str == NULL) return NULL;
    size_t len = strlen(str);
    char *reversed = malloc(len + 1);
    if (reversed == NULL) return NULL;
    for (size_t i = 0; i < len; i++) {
        reversed[i] = str[len - 1 - i];
    }
    reversed[len] = '\0';
    return reversed;
}

char *string_concat(const char *str1, const char *str2) {
    if (str1 == NULL) str1 = "";
    if (str2 == NULL) str2 = "";
    size_t len1 = strlen(str1);
    size_t len2 = strlen(str2);
    char *result = malloc(len1 + len2 + 1);
    if (result == NULL) return NULL;
    strcpy(result, str1);
    strcpy(result + len1, str2);
    return result;
}
