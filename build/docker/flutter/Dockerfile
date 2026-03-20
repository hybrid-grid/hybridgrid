# Flutter Android base image for hybridgrid worker nodes.
# Base: ghcr.io/cirruslabs/flutter (cirrusci/flutter on Docker Hub is deprecated).
# Android SDK is pre-bundled at /opt/flutter/bin/cache/android-sdk.
#
# Build args:
#   FLUTTER_VERSION  - Flutter SDK version (default: 3.19.6)
#
# Exposes:
#   FLUTTER_ROOT, ANDROID_SDK_ROOT, JAVA_HOME, PATH (Flutter/Dart/Android/Gradle bins).
#
# No iOS toolchain, no Xcode, no signing keys.

ARG FLUTTER_VERSION=3.19.6

# Stage 1: accept Android licenses
FROM ghcr.io/cirruslabs/flutter:${FLUTTER_VERSION} AS licenses

ENV DEBIAN_FRONTEND=noninteractive

# Accept Android SDK licenses non-interactively; || true so build succeeds
# even if licenses are already accepted.
RUN yes | flutter doctor --android-licenses > /dev/null 2>&1 || true

# Stage 2: final image
FROM ghcr.io/cirruslabs/flutter:${FLUTTER_VERSION}

# Copy accepted licenses from stage 1
COPY --from=licenses /root/.android /root/.android

ENV FLUTTER_ROOT=/opt/flutter
ENV ANDROID_SDK_ROOT=/opt/flutter/bin/cache/android-sdk
ENV ANDROID_HOME=/opt/flutter/bin/cache/android-sdk
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-arm64

ARG GRADLE_VERSION=8.5
RUN curl -sSL "https://services.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip" -o /tmp/gradle.zip && \
    unzip -q /tmp/gradle.zip -d /opt && \
    mv /opt/gradle-${GRADLE_VERSION} /opt/gradle && \
    rm /tmp/gradle.zip
ENV PATH="${FLUTTER_ROOT}/bin:${FLUTTER_ROOT}/bin/cache/dart-sdk/bin:${ANDROID_SDK_ROOT}/cmdline-tools/latest/bin:${ANDROID_SDK_ROOT}/platform-tools:${JAVA_HOME}/bin:/opt/gradle/bin:${PATH}"

WORKDIR /workspace

LABEL org.opencontainers.image.source="https://github.com/h3nr1-d14z/hybridgrid" \
      org.opencontainers.image.description="Flutter Android build environment for hybridgrid workers" \
      flutter.version="${FLUTTER_VERSION}"

# Keep container alive — actual worker integration in later tasks.
CMD ["tail", "-f", "/dev/null"]
