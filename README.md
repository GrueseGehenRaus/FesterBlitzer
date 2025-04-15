# FesterBlitzer

FesterBlitzer is a sleek heads-up display (HUD) interface that displays real-time vehicle data — such as speed and RPM — via OBD2. It also fetches and shows the distance to nearby speed cameras (blitzers), helping you stay alert and drive safely.
## 🚀 Getting Started

Before running the application, make sure to update the path to your OBD2 adapter (typically a USB device):


```go
device := initDevice("/dev/tty.usbserial-11340") //in my case
```
For testing purposes, you can use a mock device by setting the path to test://. This allows you to simulate OBD2 responses (see the fork of elmobd for details):

```go
device := initDevice("test://")
```
## ▶️ Run the Application

To start the app:

```go
go run main.go
```
