/**
 * Represents a disk drive or storage device.
 * @interface Disk
 * @property {string} Name - The name of the disk.
 * @property {string} Path - The file system path where the disk is mounted or located.
 * @property {string} Size - The storage capacity of the disk.
 * @property {Date} Created - The date when the disk was created or formatted.
 * @property {Date} Modified - The date when the disk was last modified.
 */
export interface Disk {
  name: string;
  path: string;
  size: string;
  created: Date;
  modified: Date;
}
