/* @refresh reload */
import { Route, Router } from '@solidjs/router';
import { render } from 'solid-js/web';
import Layout from './components/layout/layout';
import NotFound from './errors/404';
import './index.css';
import About from './pages/about';
import Home from './pages/home';
import Login from './pages/login';

render(
  () => (
      <Router>
          <Route path="/" component={Layout}>
            <Route path="/" component={Home} />
            <Route path="/about" component={About} />
          </Route>
          <Route path="/login" component={Login} />
          <Route path="**" component={NotFound} />
      </Router>
  ),
  document.getElementById('root')
);
