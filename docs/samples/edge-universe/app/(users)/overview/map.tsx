import React from "react";
import {GoogleMap, useLoadScript, Marker} from '@react-google-maps/api';
import { marker } from "leaflet";

interface MapProps extends google.maps.MapOptions {
    markers: {id:string, lat: number, lng: number}[];
    style: { [key: string]: string };
    onClick?: (e: google.maps.MapMouseEvent) => void;
    onIdle?: (map: google.maps.Map) => void;
    children?: React.ReactNode;
  }
 
const Map: React.FC<MapProps> = ({markers}) => {
    const {isLoaded} = useLoadScript({
        googleMapsApiKey: process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY,
    })
    if (!isLoaded) return <div>loading...</div>
    return (
        <GoogleMap zoom={5} center={{lat: 38, lng: -100}} mapContainerClassName="map-container">            
            {markers.map((marker, index)=>(
                <Marker key={marker.id} position={{lat:marker.lat, lng: marker.lng}} />
            ))}
            {markers.map((marker, index)=>(
                <Marker key={marker.id} position={{lat:marker.lat, lng: marker.lng}} />
            ))}                  
        </GoogleMap>       
    )
};

export default Map;