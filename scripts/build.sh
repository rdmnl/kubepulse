#!/bin/bash

# Create a directory for distribution files
mkdir -p dist

# Define target platforms as an array of strings
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"
do
    # Split platform into GOOS and GOARCH
    GOOS="${platform%%/*}"
    GOARCH="${platform##*/}"
    output_name="kubepulse"

    # Add .exe extension for Windows
    if [ "$GOOS" = "windows" ]; then
        output_name+=".exe"
    fi

    echo "Building for $GOOS $GOARCH..."

    # Set environment variables and build
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -ldflags="-s -w" -o "$output_name"
    if [ $? -ne 0 ]; then
        echo "Error occurred during build for $GOOS $GOARCH!"
        exit 1
    fi

    # Package the binary into an archive
    if [ "$GOOS" = "windows" ]; then
        zip "dist/kubepulse-$GOOS-$GOARCH.zip" "$output_name"
    else
        tar -czf "dist/kubepulse-$GOOS-$GOARCH.tar.gz" "$output_name"
    fi

    # Clean up the binary
    rm "$output_name"

    echo "Built and packaged kubepulse-$GOOS-$GOARCH"
done

echo "All builds completed successfully."


#!/bin/bash

# # Create a directory for distribution files
# mkdir -p dist

# # Define target platforms as an array of strings
# platforms=(
#     "linux/amd64"
#     "linux/arm64"
#     "darwin/amd64"
#     "darwin/arm64"
#     "windows/amd64"
# )

# for platform in "${platforms[@]}"
# do
#     # Split platform into GOOS and GOARCH
#     GOOS="${platform%%/*}"
#     GOARCH="${platform##*/}"
#     output_name="kubepulse"

#     # Add .exe extension for Windows
#     if [ "$GOOS" = "windows" ]; then
#         output_name+=".exe"
#     fi

#     echo "Building for $GOOS $GOARCH..."

#     # Set environment variables and build
#     env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -ldflags="-s -w" -o "$output_name"
#     if [ $? -ne 0 ]; then
#         echo "Error occurred during build for $GOOS $GOARCH!"
#         exit 1
#     fi

#     # For Linux builds, compress the binary using upx
#     if [ "$GOOS" = "linux" ]; then
#         echo "Compressing binary with upx for $GOOS $GOARCH..."
#         upx --best --ultra-brute "$output_name"
#         if [ $? -ne 0 ]; then
#             echo "Error occurred during UPX compression for $GOOS $GOARCH!"
#             exit 1
#         fi
#     fi

#     # Package the binary into an archive
#     if [ "$GOOS" = "windows" ]; then
#         zip "dist/kubepulse-$GOOS-$GOARCH.zip" "$output_name"
#     else
#         tar -czf "dist/kubepulse-$GOOS-$GOARCH.tar.gz" "$output_name"
#     fi

#     # Clean up the binary
#     rm "$output_name"

#     echo "Built and packaged kubepulse-$GOOS-$GOARCH"
# done

# echo "All builds completed successfully."
