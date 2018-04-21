## Ballistics calculator for long range shooting written (somewhat poorly for now) in Go

I built this as a learning project for golang, its a functional ballistics calculator 
that outputs distances per milliradian of scope dope based on bullet characteristics
and scope zero distance. its pretty solid supersonic, subsonic gets a little wonkey 
with drag coefficient.  next time i touch this i will calculate cp dynamically vs velocity 
in the subsonic range. 


also i need to rewrite it in a more sensible fashion now that i know go slightly 
better
