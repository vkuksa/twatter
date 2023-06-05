#!/bin/bash

# Install Git
echo "Installing Git..."
sudo apt-get update
sudo apt-get install git -y

# Clone the repository
echo "Cloning the repository..."
git clone https://github.com/vkuksa/twatter.git

# Change to the repository directory
cd twatter

# Execute "make run"
echo "Executing 'make run'..."
make
