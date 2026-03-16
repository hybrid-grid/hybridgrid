#ifndef STRING_UTILS_H
#define STRING_UTILS_H

#include <stddef.h>

size_t string_length(const char *str);
char *string_reverse(const char *str);
char *string_concat(const char *str1, const char *str2);

#endif
