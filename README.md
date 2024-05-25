Overview

The Pod Cleanup Operator is a Kubernetes operator designed to manage and clean up 
pods in specific failure states, such as evicted pods, crashloopbackoff pods, image pull 
error pods, and failed pods. This helps maintain a clean and efficient Kubernetes 
environment by automatically deleting pods that meet the specified criteria.



Features

●Evicted Pods Cleanup: Deletes pods that have been evicted.
●CrashLoopBackOff Pods Cleanup: Deletes pods stuck in a CrashLoopBackOff 
state.
●ImagePullError Pods Cleanup: Deletes pods that have failed due to image pull 
errors.
●Failed Pods Cleanup: Deletes pods that have reached a failed state.
