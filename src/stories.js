// Storymap registry — every JSON in the project root is a selectable story.
// The user picks one in the top-right dropdown; the app remounts per story.
import freesoundOcean from '../freesound-ocean-storymap.json';
import pixabayImages from '../pixabay-image-storymap.json';
import pixabayVideos from '../pixabay-video-storymap.json';
import surfing from '../surfing-storymap.json';

export const stories = [
    { id: 'freesound-ocean', label: 'Freesound · Ocean recordings', config: freesoundOcean },
    { id: 'pixabay-images', label: 'Pixabay · Ocean images', config: pixabayImages },
    { id: 'pixabay-videos', label: 'Pixabay · Ocean videos', config: pixabayVideos },
    { id: 'surfing', label: 'Pixabay · Surfing videos', config: surfing },
];

export const defaultStoryId = 'freesound-ocean';

export function getStory(id) {
    return stories.find((s) => s.id === id) || stories[0];
}
