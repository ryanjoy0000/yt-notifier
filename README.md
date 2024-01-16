# Youtube Notifier (Telegram Bot) - A real time data processing pipeline
**TL:DR** : This app helps in tracking changes to someone else's YouTube video/s.

## Project Description

### What your application does?

This project is an implementation of a generic real time data processing pipeline  It helps in tracking the changes in the YouTube statistics (likes, comments, etc.) of a video (or list of videos) not owned by me on YouTube. A telegram bot at the end of the pipeline provides the relevant info as alerts.

### Technologies you used

Go, Apache Kafka, Confluent Cloud, ksqlDB, Telegram API, Google API

### What problem does it solve? Why did you build this project?

This question has long bothered me: How can you sign up for notifications from a system that doesn't offer an alerts API? How can we get a computer or system to monitor something we want to be aware of on the internet and notify us when it changes? For instance, let's say that my friends and I participate in a video that is recorded and posted to YouTube. Since I don't own the video, YouTube won't notify me when there are any new likes or comments.

The strategy used here is that I start by creating a playlist of things I want to watch on YouTube, and every time I want to watch a new video, I can just add it to that playlist. Look up the videos using a script in Go, check their statistics with Google API, and just stream that snapshot of data up to Kafka and deal with it by stream processing. I can then pull out the relevant info into my phone using Telegram API and a Kafka Connector. (Disclaimer: The solution is not originally mine. It is adapted from a python solution found in the internet. )

### What can be improved?
Project is at a very rudimentary stage but functional. A more micro-service oriented architecture will help it to be scalable. For efficiency, code refactoring is required. Deployment process can be improved with Docker and scheduler systems.

## Installation instructions
- Download and extract the zip file to a folder in the local machine.
- Get the keys for YouTube API & Telegram API Key
- Create appropriate config files  with the above info along with kafka properties 
- Run the command in CMD (on the folder path): `make run`

## Demo
