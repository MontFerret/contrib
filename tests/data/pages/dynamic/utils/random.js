export default function random(min = 1000, max = 4000) {
    if (max <= min) {
        return min;
    }

    return Math.floor(min + Math.random() * (max - min + 1));
}
