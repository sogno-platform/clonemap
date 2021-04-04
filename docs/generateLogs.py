import json
import numpy as np
import datetime


curr = datetime.datetime.now()
topics = ["error", "debug", "msg", "status", "app"]
data = []
for x in range(0, 10):
   for topic in topics:
      # select the agentid randomly from 0 to 9
      agentid = np.random.randint(0, 5)
      interval = np.random.randint(1, 60 * 60 * 24 * 30)
      timestamp = curr - datetime.timedelta(seconds=interval)

      #topic = np.random.choice(topics)
      logMessage = {
         "masid": 0,
         "agentid": agentid,
         "timestamp": timestamp,
         "topic": topic,
         "msg": "msg",
         "data": ""
      }



      if topic == "msg":
         logMessage["msg"] = "ACL send"
         agentids = [0, 1, 2, 3, 4]
         agentids.remove(agentid)
         receiver = np.random.choice(agentids)
         logMessage["data"] = "Sender: " + str(agentid) + ";Receiver: " + str(receiver) + ";Timestamp: "  + logMessage["timestamp"].strftime("%Y-%m%dT%H:%M:%S")
         logMessageReceiver = {
            "masid": 0,
            "agentid": int(receiver),
            "timestamp": timestamp,
            "topic": topic,
            "msg": "ACL receive",
            "data": "Sender: " + str(agentid) + ";Receiver: " + str(receiver) + ";Timestamp: " +logMessage["timestamp"].strftime("%Y-%m-%dT%H:%M:%S")
         } 

         data.append(logMessage)
         data.append(logMessageReceiver)
      
      else:
         data.append(logMessage)

data = sorted(data, key=lambda k: k["timestamp"], reverse=True)
for log in data:
   log["timestamp"] = log["timestamp"].strftime('%Y-%m-%dT%H:%M:%S')

with open("logs.json", "w") as write_file:
    json.dump(data, write_file) 