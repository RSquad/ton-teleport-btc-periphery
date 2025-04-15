cmake_minimum_required(VERSION 3.16)

project(test LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED True)

find_package(PkgConfig REQUIRED)
pkg_check_modules(LIBSODIUM REQUIRED libsodium)

include_directories(${LIBSODIUM_INCLUDE_DIRS})

add_executable(test main.cpp)

set(ED2X_PATH ./../build/libed25519_to_x25519.a)
target_link_libraries(test PRIVATE ${LIBSODIUM_LIBRARIES} ${ED2X_PATH})

install(TARGETS test
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
)
