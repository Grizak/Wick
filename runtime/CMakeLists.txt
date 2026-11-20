cmake_minimum_required(VERSION 3.20)
project(wick_runtime C)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

set(WICK_RUNTIME_SOURCES 
    src/runtime.c
    src/args/args.c
)

set(WICK_RUNTIME_HEADERS
    src/runtime.h
    src/args/args.h
)

add_library(wick_runtime STATIC ${WICK_RUNTIME_SOURCES})

target_include_directories(wick_runtime PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/src>
    $<INSTALL_INTERFACE:include>
)

set_target_properties(wick_runtime PROPERTIES POSITION_INDEPENDENT_CODE ON)

if(NOT MSVC)
    target_compile_options(wick_runtime PRIVATE -Wall -Wextra -O2)
endif()

install(TARGETS wick_runtime ARCHIVE DESTINATION lib)
install(FILES ${WICK_RUNTIME_HEADERS} DESTINATION include)
