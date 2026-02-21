export function checkLoggedIn(data?: { jwt?: string }): boolean {
	return !!data?.jwt;
}
