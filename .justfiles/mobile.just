ios_device := 'BB814B76-09C2-4269-8ACF-930040C43EFD'

[working-directory('frontend/mobile')]
mobile:
    bun run dev --open

[working-directory('frontend/mobile')]
build-mobile: install-mobile-deps
    bun run build

[working-directory('frontend/mobile')]
run-mobile-ios: build-mobile
    bun cap run --target {{ ios_device }} ios

[working-directory('frontend/mobile')]
run-mobile-android: build-mobile
    ADB_INSTALL_TIMEOUT=60 bun cap run android

[working-directory('frontend/mobile')]
open-mobile-ios: build-mobile
    xcrun simctl shutdown all
    xcrun simctl erase {{ ios_device }}
    bun cap open ios

[working-directory('frontend/mobile')]
open-mobile-android: build-mobile
    bun cap open android
