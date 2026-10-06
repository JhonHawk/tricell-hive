// @ts-check
import starlight from "@astrojs/starlight";
import { defineConfig } from "astro/config";

export default defineConfig({
	site: "https://hive.tricell.tech",
	integrations: [
		starlight({
			title: "Hive",
			logo: {
				dark: "./src/assets/hive-mark.svg",
				light: "./src/assets/hive-mark-light.svg",
				alt: "",
				replacesTitle: false,
			},
			customCss: ["./src/styles/theme.css"],
			sidebar: [
				{ label: "Start", items: [{ slug: "index" }] },
				{
					label: "Flows",
					items: [
						{ slug: "flows/research" },
						{ slug: "flows/plan" },
						{ slug: "flows/build" },
						{ slug: "flows/close" },
					],
				},
			],
		}),
	],
});
