#include "smoke.h"

#include <utility>

template <int N>
struct Fibonacci {
    static constexpr int value = Fibonacci<N - 1>::value + Fibonacci<N - 2>::value;
};

template <>
struct Fibonacci<1> { static constexpr int value = 1; };

template <>
struct Fibonacci<0> { static constexpr int value = 0; };

template <std::size_t... I>
constexpr int expanded_sum(std::index_sequence<I...>) {
    return (0 + ... + static_cast<int>(I % 7));
}

int template_value() {
    constexpr auto indices = std::make_index_sequence<256>{};
    return Fibonacci<20>::value % 10 + expanded_sum(indices) - expanded_sum(indices);
}
