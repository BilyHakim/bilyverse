import { mount } from 'svelte'
import '@fontsource-variable/manrope/wght.css'
import '@fontsource/ibm-plex-mono/latin-400.css'
import './style.css'
import App from './App.svelte'

mount(App, { target: document.getElementById('app') })
