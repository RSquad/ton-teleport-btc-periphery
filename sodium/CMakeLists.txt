cmake_minimum_required(VERSION 3.16)

project(ed25519_to_x25519 LANGUAGES C)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED True)

find_package(PkgConfig REQUIRED)
pkg_check_modules(LIBSODIUM REQUIRED libsodium)

include_directories(${LIBSODIUM_INCLUDE_DIRS})

add_library(ed25519_to_x25519 STATIC x25519_crypto.c)

target_link_libraries(ed25519_to_x25519 PRIVATE ${LIBSODIUM_LIBRARIES})

install(TARGETS ed25519_to_x25519
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
)
