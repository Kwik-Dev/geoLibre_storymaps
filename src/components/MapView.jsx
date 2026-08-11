import React, { useEffect, useRef } from 'react';
import maplibregl from 'maplibre-gl';

/**
 * Fixed full-viewport MapLibre map with an optional inset overview map and a
 * location marker. Exposes the live map/marker instances through refs so the
 * scroll controller (App) can drive the camera imperatively.
 */
export default function MapView({ config, mapRef, insetMapRef, markerRef, insetMarkerRef, onReady }) {
    const mapContainerRef = useRef(null);
    const insetContainerRef = useRef(null);
    const onReadyRef = useRef(onReady);
    onReadyRef.current = onReady;

    useEffect(() => {
        // Shape right-to-left scripts (Arabic, Hebrew, Persian, …) correctly so
        // basemap labels are not rendered reversed. Lazy-loaded, so it only
        // downloads when an RTL label is actually encountered. Pinned version
        // (no SRI possible for worker-imported scripts), matching the original.
        if (maplibregl.getRTLTextPluginStatus?.() === 'unavailable') {
            maplibregl
                .setRTLTextPlugin('https://unpkg.com/@mapbox/mapbox-gl-rtl-text@0.4.0/dist/mapbox-gl-rtl-text.js', true)
                .catch((e) => console.error('[GeoLibre] RTL plugin failed', e));
        }

        // For a global start slide, open at the world view rather than chapter 0
        // so a slow style/tile load doesn't flash chapter 0's camera.
        const openLocation = config.startSlide === 'global' ? config.globalView : config.chapters[0].location;

        const map = new maplibregl.Map({
            container: mapContainerRef.current,
            style: config.style,
            center: openLocation.center,
            zoom: openLocation.zoom,
            bearing: openLocation.bearing,
            pitch: openLocation.pitch,
            interactive: false,
        });
        mapRef.current = map;

        if (config.inset) {
            const inset = new maplibregl.Map({
                container: insetContainerRef.current,
                style: config.insetStyle,
                center: openLocation.center,
                zoom: config.insetZoom || 1,
                interactive: false,
                attributionControl: false,
            });
            insetMapRef.current = inset;
            const markerEl = document.createElement('div');
            markerEl.className = 'inset-marker';
            insetMarkerRef.current = new maplibregl.Marker({ element: markerEl })
                .setLngLat(openLocation.center)
                .addTo(inset);
            // The global overview shows no marker; hide it immediately.
            if (config.startSlide === 'global') markerEl.style.visibility = 'hidden';
        }

        if (config.showMarkers) {
            const marker = new maplibregl.Marker({ color: config.markerColor })
                .setLngLat(openLocation.center)
                .addTo(map);
            markerRef.current = marker;
            if (config.startSlide === 'global') marker.getElement().style.visibility = 'hidden';
        }

        map.once('load', () => {
            // Match the in-app projection (globe by default) so the story does
            // not silently fall back to 2D Mercator (#917).
            try {
                map.setProjection({ type: config.projection || 'globe' });
            } catch (e) {
                console.error('[GeoLibre] projection failed', e);
            }
            onReadyRef.current();
        });

        return () => {
            map.remove();
            if (insetMapRef.current) {
                insetMapRef.current.remove();
                insetMapRef.current = null;
            }
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return (
        <>
            <div id="map" ref={mapContainerRef} />
            {config.inset && (
                <div id="inset-map" ref={insetContainerRef} className={config.insetPosition || 'bottom-left'} />
            )}
        </>
    );
}
