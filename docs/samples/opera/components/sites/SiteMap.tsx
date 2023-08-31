'use client';

import { Site } from '../../app/types';
import {useState} from 'react';
import {GoogleMap, InfoWindowF, MarkerF, useJsApiLoader} from '@react-google-maps/api';
interface SiteMapProps {
    sites: Site[];
}

function SiteMap(props: SiteMapProps) {
    const { sites } = props;
    const { isLoaded } = useJsApiLoader({
        id: 'google-map-script',
        googleMapsApiKey: process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY ?? '',
    });
    const [map, setMap] = useState(null);
    const [selected, setSelected] = useState<Site | null>(null);
    const [center, setCenter] = useState({lat: 0, lng: 0});
    const [zoom, setZoom] = useState(2);
    const containerStyle = {
        display: 'flex',
        width: '100%',
        height: '800px'
    };
    if (!isLoaded) return <div className="text-7xl text-green-500">Loading...</div>;
    const onMarkerClick = (site: any) => {
        setSelected(site);
    };
    const onInfoWindowClose = () => {
        setSelected(null);
    };
    const onMapClick = () => {
        setSelected(null);
    };
    
    return <div
    style={{
      display: "flex",
      flexDirection: "column",
      justifyContent: "center",
      alignItems: "center",
        width: "100%",
        height: "100%",
      backgroundColor: "red",
    }}
  >
        <GoogleMap
                mapContainerStyle={containerStyle}
                center={center}
                zoom={zoom}
                mapContainerClassName="map"
                // onLoad={onLoad}
                // onUnmount={onUnmount}
                // onClick={onMapClick}
            >
                {sites.map((site: Site) => (
                    <MarkerF
                        key={site.id}
                        position={{lat: site.lat, lng: site.lng}}
                        onClick={() => onMarkerClick(site)}
                    />
                ))}
                {selected && (
                    <InfoWindowF
                        position={{lat: selected.lat, lng: selected.lng}}
                        onCloseClick={onInfoWindowClose}
                    >
                        <div>
                            <h2>{selected.name}</h2>
                            <p>{selected.address}</p>
                        </div>
                    </InfoWindowF>
                )}
            </GoogleMap>
            </div>;
}
export default SiteMap;