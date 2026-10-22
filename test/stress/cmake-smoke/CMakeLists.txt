cmake_minimum_required(VERSION 3.16)
project(hg_cmake_smoke LANGUAGES C CXX)

set(CMAKE_C_STANDARD 11)
set(CMAKE_CXX_STANDARD 17)

add_executable(hg_cmake_smoke
    main.cc
    part1.cc
    part2.cc
    part3.cc
    part4.cc
    template.cc
    helper.c
)
target_include_directories(hg_cmake_smoke PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
