# HolidayParksAPI
rest api for HolidayParks 


router.GET/reservations/licensePlate/:licensePlate"
this will return a true or false bool. it is used to check if a licensePlate is in the database or not

router.POST"/reservation"
this is used to creat a reservation
{ 
    "firstName": "your_firstName",
    "lastName": "your_lastName",
    "phoneNumber": "your_phoneNumber",
    "licensePlate": "your_licensePlate",
    "dateOfDeparture": "your_dateOfDeparture",
    "dateOfArrival": "your_dateOfArrival"
}

router.PATCH"/reservations/:reservation_id"
this is used to change a reservation.
{ 
    "firstName": "your_firstName",
    "lastName": "your_lastName",
    "phoneNumber": "your_phoneNumber",
    "licensePlate": "your_licensePlate",
    "dateOfDeparture": "your_dateOfDeparture",
    "dateOfArrival": "your_dateOfArrival"
}

router.DELETE"/reservations/:reservation_id"
this is used to delete a reservation. 