// Username generation and storage utilities

const ADJECTIVES = [
  'Happy', 'Lucky', 'Swift', 'Brave', 'Clever', 'Gentle', 'Mighty', 'Noble',
  'Quiet', 'Royal', 'Silent', 'Witty', 'Cosmic', 'Mystic', 'Golden', 'Silver',
  'Crystal', 'Shadow', 'Bright', 'Wild', 'Cool', 'Epic', 'Rad', 'Chill'
];

const NOUNS = [
  'Tiger', 'Eagle', 'Dragon', 'Phoenix', 'Wolf', 'Bear', 'Fox', 'Hawk',
  'Lion', 'Panther', 'Falcon', 'Raven', 'Otter', 'Panda', 'Koala', 'Dolphin',
  'Wizard', 'Knight', 'Ninja', 'Pirate', 'Viking', 'Samurai', 'Ranger', 'Mage'
];

const USERNAME_STORAGE_KEY = 'broadcast_box_username';
const USER_ID_STORAGE_KEY = 'broadcast_box_user_id';

/**
 * Generates a random username like "HappyTiger123"
 */
export function generateRandomUsername(): string {
  const adjective = ADJECTIVES[Math.floor(Math.random() * ADJECTIVES.length)];
  const noun = NOUNS[Math.floor(Math.random() * NOUNS.length)];
  const number = Math.floor(Math.random() * 1000);

  return `${adjective}${noun}${number}`;
}

/**
 * Gets the stored username or generates a new one
 */
export function getUsername(): string {
  try {
    const stored = localStorage.getItem(USERNAME_STORAGE_KEY);
    if (stored) {
      return stored;
    }
  } catch (e) {
    console.error('Failed to access localStorage:', e);
  }

  // Generate and store new username
  const newUsername = generateRandomUsername();
  setUsername(newUsername);
  return newUsername;
}

/**
 * Sets the username in localStorage
 */
export function setUsername(username: string): void {
  try {
    localStorage.setItem(USERNAME_STORAGE_KEY, username);
  } catch (e) {
    console.error('Failed to save username:', e);
  }
}

/**
 * Gets or generates a persistent user ID
 */
export function getUserId(): string {
  try {
    const stored = localStorage.getItem(USER_ID_STORAGE_KEY);
    if (stored) {
      return stored;
    }
  } catch (e) {
    console.error('Failed to access localStorage:', e);
  }

  // Generate new user ID (simple UUID v4)
  const newId = generateUUID();
  try {
    localStorage.setItem(USER_ID_STORAGE_KEY, newId);
  } catch (e) {
    console.error('Failed to save user ID:', e);
  }

  return newId;
}

/**
 * Simple UUID v4 generator
 */
function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/**
 * Validates username format (3-20 alphanumeric characters)
 */
export function isValidUsername(username: string): boolean {
  if (!username || username.length < 3 || username.length > 20) {
    return false;
  }

  // Allow alphanumeric and underscores
  return /^[a-zA-Z0-9_]+$/.test(username);
}

/**
 * Clears stored username (for testing)
 */
export function clearUsername(): void {
  try {
    localStorage.removeItem(USERNAME_STORAGE_KEY);
  } catch (e) {
    console.error('Failed to clear username:', e);
  }
}
