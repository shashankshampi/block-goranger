# goranger
### **Performance and Quality**

VERTICAL PERF STEPS

A.  HOW TO ONBOARD AN API TO goranger

     Send a PR branch or upload it on the fly with the following details

1. Create a profile json with list of tasks/api which needs to be load tested  
2. Data or CSV files should be under files/<pod_name>/filename.csv 
3. Respective API payload (if any) should be under payloads/<pod-name>/payload.json 

B. HOW TO WRITE PLACE HOLDERS FOR THE REPLACING FROM CSV DATA?

 Replacer should be like below. It can be used anywhere in the payload, url and headers
1. Should start with dollar sign `$` 
2. Should be enclosed between flower braces 
3. CSV name and value index should be separated by underscore  `_`  

How To RUN:

`go run main.go -profile profiles/profilename.json`