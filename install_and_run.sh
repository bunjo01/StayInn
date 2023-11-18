#!/bin/bash

# Navigate to the StayInn_Front folder
cd StayInn_Front || exit

# Install Node modules
npm install

# Run the app with npm start and navigate back to the root of the project
npm start & cd ..

# Build and start the dockerized system with docker-compose
docker-compose build --no-cache && docker-compose up
