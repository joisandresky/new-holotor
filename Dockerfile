FROM golang:1.8

# Create the directory where the application will reside
RUN mkdir /app

# Copy the application files (needed for production)
ADD new-holotor /app/new-holotor
ADD controllers /app/controllers
ADD models /app/models
ADD config /app/config

# Set the working directory to the app directory
WORKDIR /app

# Expose the application on port 8080.
# This should be the same as in the app.conf file
EXPOSE 8989

# Set the entry point of the container to the application executable
ENTRYPOINT /app/new-holotor