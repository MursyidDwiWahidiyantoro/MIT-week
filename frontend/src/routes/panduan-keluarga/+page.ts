import { redirect } from '@sveltejs/kit';

// The family guide now lives inside Belajar. The old address keeps working.
export function load(): never {
  redirect(308, '/edukasi#keluarga');
}
