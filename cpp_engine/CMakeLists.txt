cmake_minimum_required(VERSION 3.10)
project(nvr_worker)

set(CMAKE_CXX_STANDARD 17)


# Find FFmpeg Libraries
find_package(PkgConfig REQUIRED)
pkg_check_modules(AVCODEC REQUIRED libavcodec)
pkg_check_modules(AVFORMAT REQUIRED libavformat)
pkg_check_modules(AVUTIL REQUIRED libavutil)

include_directories(${AVCODEC_INCLUDE_DIRS} ${AVFORMAT_INCLUDE_DIRS} ${AVUTIL_INCLUDE_DIRS})
link_directories(${AVCODEC_LIBRARY_DIRS} ${AVFORMAT_LIBRARY_DIRS} ${AVUTIL_LIBRARY_DIRS})

# Set cpp compile list
file(GLOB_RECURSE VI_CPP "vi/*.cpp")
file(GLOB_RECURSE LOG_CPP "log/*.cpp")

add_executable(nvr_worker main.cpp ${VI_CPP} ${LOG_CPP})

# Header search folders
target_include_directories(nvr_worker PRIVATE ${CMAKE_CURRENT_SOURCE_DIR}
                        ${CMAKE_CURRENT_SOURCE_DIR}/vi
                        ${CMAKE_CURRENT_SOURCE_DIR}/log
                        )

target_link_libraries(nvr_worker
    ${AVCODEC_LIBRARIES}
    ${AVFORMAT_LIBRARIES}
    ${AVUTIL_LIBRARIES}
    pthread
)

