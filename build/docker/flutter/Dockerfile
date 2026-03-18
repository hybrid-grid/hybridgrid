# Flutter Android base image for hybridgrid worker nodes
# Base: ghcr.io/cirruslabs/flutter (NOT cirrusci/flutter which is deprecated)
# Android SDK is bundled in the cirruslabs Flutter images on non-macOS hosts
#
# Build args:
#   FLUTTER_VERSION  - Flutter SDK version (default: 3.19.6)
#   ANDROID_SDK_TAG - Android SDK component tag (default: 11076708)
#
# Exposes:
#   FLUTTER_ROOT, ANDROID_SDK_ROOT, JAVA_HOME, PATH (Flutter/Dart bins)
#
# No iOS toolchain, no Xcode, no signing keys.

ARG FLUTTER_VERSION=3.19.6

# Flutter SDK with Android tooling pre-bundled (debian variant has best toolchain support)
FROM ghcr.io/cirruslabs/flutter:${FLUTTER_VERSION} AS base

# Prevent interactive prompts during apt operations
ENV DEBIAN_FRONTEND=noninteractive

# ---- Stage 1: Accept Android licenses ----
FROM base AS licenses

# Accept all Android SDK licenses non-interactively
RUN yes | ${FLUTTER_ROOT}/bin/flutter doctor --android-licenses 2>/dev/null || true

# ---- Stage 2: Final image ----
FROM base AS builder

# Copy accepted licenses from licenses stage
COPY --from=licenses /root/.android /root/.android
COPY --from=licenses /root/.android /home/flutter/.android

# Additional Android SDK components can be installed here if needed:
# RUN ${FLUTTER_ROOT}/bin/flutter config --android-sdk /opt/android-sdk && \
#     sdkmanager --install "platform-tools" "build-tools;34.0.0" "platforms;android-34"

# ---- Final image ----
FROM base

# Copy accepted licenses into final image
COPY --from=licenses /root/.android /root/.android
COPY --from=licenses /root/.android /home/flutter/.android

# Environment variables
ENV FLUTTER_ROOT=/opt/flutter
ENV ANDROID_SDK_ROOT=${FLUTTER_ROOT}/bin/cache/android-sdk
ENV ANDROID_HOME=${ANDROID_SDK_ROOT}
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64

# Ensure Flutter and Dart binaries are on PATH
ENV PATH="${FLUTTER_ROOT}/bin:${FLUTTER_ROOT}/bin/cache/dart-sdk/bin:${ANDROID_SDK_ROOT}/cmdline-tools/latest/bin:${ANDROID_SDK_ROOT}/platform-tools:${JAVA_HOME}/bin:${PATH}"

RUN flutter --version && \
    echo "ANDROID_SDK_ROOT=${ANDROID_SDK_ROOT}" && \
    flutter doctor

# Default command: keep container alive with a simple no-op
CMD ["tail", "-f", "/dev/null"]
