function started_less_than_seconds_ago(event, seconds) {
	var seconds = seconds || 600;
    try {
        if (event.entity.hasOwnProperty("annotations")) {
    		if (event.entity.hasOwnProperty("started_at")) {
        		var ts = (new Date).getTime()/1000;
        		res = (ts - event.started_at) < seconds
                return res
        	}
        }
    }
    catch(err) {
        console.log("Failed to get entity annotations:");
        console.log(err.message);
        return false;
    }
	return false
}

/*
event = {"entity":{"annotations":{"started_at":1603902240}}}
started_more_than_seconds_ago(event, 1800);
*/
